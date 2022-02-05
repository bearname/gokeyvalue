package main

import (
    "bytes"
    "encoding/json"
    "errors"
    "flag"
    "github.com/gorilla/mux"
    "gokeyvalue/pkg/model"
    "gokeyvalue/pkg/repo"
    "gokeyvalue/pkg/service"
    "gokeyvalue/pkg/transport"
    "gokeyvalue/pkg/util/concurrent"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "sync"
    "syscall"
    "time"
)

var ErrServerDown = errors.New("server not available")

func randInRange(r *rand.Rand, min, max int) int {
    return r.Intn(min) + (max - min)
}

func main() {
    var isMaster bool
    var port int
    var instanceId int
    var otherServer string
    var volume string
    var guarantees int
    flag.StringVar(&volume, "volume", "./", "volume path")
    flag.BoolVar(&isMaster, "isMaster", false, "is master server")
    flag.IntVar(&port, "port", 8000, "port")
    flag.IntVar(&instanceId, "instanceId", 8000, "port")
    flag.StringVar(&otherServer, "urls", "", " -urls=1!192.168.0.2:8000,2!192.168.0.3:8001,")
    flag.IntVar(&guarantees, "guarantees", int(model.GuaranteesAsync), "async=0, sync=1")
    flag.Parse()

    seen := make(map[string]bool)
    flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
    required := []string{"instanceId", "port"}
    for _, req := range required {
        if !seen[req] {
            // or possibly use `log.Fatalf` instead of:
            log.Fatal("missing required -" + req + "root argument/flag\n")
        }
    }
    parseGuarantees, err := model.ParseGuarantees(guarantees)
    if err != nil {
        log.Fatal(err)
    }

    urls := strings.Split(otherServer, ",")
    var replicaList []model.Server
    if len(urls) == 0 {
        isMaster = true
    }

    //connectorService := service.NewConnectorService(0)

    if len(otherServer) > 0 {
        var server model.Server
        for _, url := range urls {
            if len(url) == 0 {
                continue
            }
            split := strings.Split(url, "!")
            id, err := strconv.Atoi(split[0])
            if err != nil {
                log.Println(errors.New("invalid rep urls format"))
                return
            }
            instId := split[1]
            if instanceId == id {
                server = *model.NewServer(id, instId, true)
            } else {
                server = *model.NewServer(id, instId, false)
            }

            replicaList = append(replicaList, server)
            //connectorService.AddOrUpdateClientDefault(i)
        }
    }
    err = os.Mkdir(volume, 0666)
    pwd, _ := os.Getwd()
    root := pwd + string(os.PathSeparator) + volume
    dataFile, err := os.OpenFile(root+"db.bin", os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        log.Fatal(err)
    }
    defer dataFile.Close()
    hashFile, err := os.OpenFile(root+"hash.txt", os.O_CREATE|os.O_RDWR, 0666)
    if err != nil {
        log.Fatal(err)
    }
    defer dataFile.Close()
    fileRepo := repo.NewFileRepo(dataFile, hashFile)
    dataService := service.NewDataService(fileRepo)

    //countNotifier := runtime.NumCPU()
    countNotifier := len(replicaList)

    //notifyChan := make(chan model.NotifyServerEvent, countNotifier)

    err = dataService.RestoreHashMap()
    if err != nil {
        log.Fatal(err)
    }
    go func() {
       for {
           dataService.BackupHashMap()
           time.Sleep(10 * time.Second)
       }
    }()

    //m := bx.New()
    //m.Add("/", cntrl.getHandler).Method(http.MethodGet)
    //m.Add("/", cntrl.setHandler).Method(http.MethodPost)
    //m.Test()
    //m.Start(":8000")

    mx := sync.Mutex{}
    wg := sync.WaitGroup{}
    hashMap := concurrent.NewConcurrentList(countNotifier)
    for i := 0; i < countNotifier - 1; i++ {
        events := make(chan model.NotifyServerEvent, 2)
        //fmt.Println(reflect.TypeOf(events))
        hashMap.Set(i, &events)
        for j := 0; j < countNotifier; j++ {
            wg.Add(1)
            go notifierSubscriber(&wg,  events, &mx, &replicaList, parseGuarantees)
        }
    }

    //for i := 0; i < countNotifier; i++ {
    //    wg.Add(1)
    //    go notifierSubscriber(&wg, i, connectorService, notifyChan, &mx, &replicaList, parseGuarantees)
    //    //go func() {
    //    //    defer wg.Done()
    //    //    //client := transport.HttpClient(countNotifier)
    //    //
    //    //    for event := range notifyChan {
    //    //        client, err := connectorService.GetClient(event.InstanceId)
    //    //        if err == nil {
    //    //            connectorService.AddOrUpdateClientDefault(event.InstanceId)
    //    //        }
    //    //
    //    //        //fmt.Println(event)
    //    //        wg.Add(1)
    //    //        notify(client, &mx, &wg, event, notifyChan, &replicaList, parseGuarantees)
    //    //    }
    //    //}()
    //}
    var cw service.ConnectionWatcher

    cntrl := newController(dataService, isMaster, replicaList, instanceId, hashMap, &cw)

    router := mux.Router{}
    router.HandleFunc("/health", cntrl.healthHandler).Methods(http.MethodGet)
    router.HandleFunc("/", cntrl.getHandler).Methods(http.MethodGet)
    router.HandleFunc("/", cntrl.setHandler).Methods(http.MethodPost)
    //router.HandleFunc("/notify", maxClients(cntrl.notifyHandler, 200)).Methods(http.MethodPost)
    router.HandleFunc("/notify", cntrl.notifyHandler).Methods(http.MethodPost)

    addr := "localhost:" + strconv.Itoa(port)
    log.Println("started on", addr)
    s := &http.Server{
        ConnState: cw.OnStateChange,
        Addr:      addr, Handler: &router,
    }
    err = s.ListenAndServe()
    if err != nil {
        log.Println(err)
    }
    //err = http.ListenAndServe("localhost:"+strconv.Itoa(port), &router)
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
}

func notifierSubscriber(wg *sync.WaitGroup, notifyChan chan model.NotifyServerEvent, mx *sync.Mutex, replicaList *[]model.Server, parseGuarantees model.ReplicationGuarantees, ) {
    defer wg.Done()
    client := transport.HttpClient(20)
    //for ind := range *replicaList {
    //   connectorService.AddOrUpdateClientDefault(strconv.Itoa(i) + "-" + strconv.Itoa(ind))
    //}
    var err error
    for event := range notifyChan {
        //now := time.Now()
        //id := strconv.Itoa(i) + "-" + strconv.Itoa(event.InstanceId)
        //client, err = connectorService.GetClient(id)
        //if err == nil {
        //   client = connectorService.AddOrUpdateClientDefault(id)
        //}
        //end := time.Since(now).Nanoseconds()
        //fmt.Println(end)

        //fmt.Println(event)
        //wg.Add(1)
        err = notify(client, mx, event, notifyChan, replicaList, parseGuarantees)
        if err == model.ErrServerDown {
            client = transport.HttpClient(10)
            //client = connectorService.AddOrUpdateClientDefault(id)
        }
    }
}

func maxClients(next http.HandlerFunc, n int) http.HandlerFunc {
    sema := make(chan struct{}, n)
    return func(w http.ResponseWriter, req *http.Request) {
        sema <- struct{}{}
        defer func() { <-sema }()

        next(w, req)
    }
}

//
//func maxClients(h http.Handler, n int) http.Handler {
//    sema := make(chan struct{}, n)
//
//    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//        sema <- struct{}{}
//        defer func() { <-sema }()
//
//        h.ServeHTTP(w, r)
//    })
//}

func notify(client *http.Client, m *sync.Mutex, event model.NotifyServerEvent, notifyChan chan model.NotifyServerEvent, replicaList *[]model.Server, guarantees model.ReplicationGuarantees) error {
    //defer wg.Done()
    replica := event.Replica
    //fmt.Println("replica", replica)
    if event.Replica.IsMaster {
        return nil
    }
    err := notifyReplicaCoordinate(client, &replica, &event, &guarantees)
    //msg := "success"
    if err != nil {
        m.Lock()
        (*replicaList)[event.ReplicaIndex].Status = model.ServerDown
        m.Unlock()
        //msg = "failed"
        //fmt.Println("notify", replica.Url, event, msg)
        time.Sleep(2 * time.Second)

        //fmt.Println("read add")
        notifyChan <- event
        return err
        //fmt.Println("read addss")
        //fmt.Println(sdf)
        //} else {
        //fmt.Println("notify", replica.Url, event, msg)
    }
    return nil
}

func notifyReplicaCoordinate(client *http.Client, replica *model.Server, event *model.NotifyServerEvent, guarantees *model.ReplicationGuarantees) error {
    if replica.IsMaster {
        return nil
    }
    //fmt.Println(guarantees.String())
    if *guarantees == model.GuaranteesSync {
        err := pingReplica(client, replica)
        if err != nil {
            return err
        }
    }

    return notifyReplica(client, replica, event, guarantees)
}

func notifyReplica(client *http.Client, replica *model.Server, value *model.NotifyServerEvent, guarantees *model.ReplicationGuarantees) error {
    marshal, err := json.Marshal(value)
    if err != nil {
        log.Println(err)
        return model.ErrFailedNotify
    }

    buffer := bytes.NewBuffer(marshal)
    url := "http://" + replica.Url + "/notify"
    switch *guarantees {
    case model.GuaranteesSync:
        //code, err := transport.SendRequest(client,http.MethodPost, url, buffer)
        resp, err := http.Post(url, "application/json; charset: utf-8;", buffer)
        if err != nil {
            log.Println(err)
            return model.ErrFailedNotify
        }
        defer resp.Body.Close()
        //
        code := resp.StatusCode
        if isSuccessNotified(code) {
            return model.ErrFailedNotify
        }
    case model.GuaranteesAsync:
        _, err = transport.SendRequest(client, http.MethodPost, url, buffer)
        //_, err = http.Post(url, "application/json; charset: utf-8;", buffer)
        if err != nil {
            //log.Println(err)
            s := err.Error()
            //fmt.Println(s)
            isDown := strings.Contains(s, "No connection could be made because the target machine actively refused it.") ||
                strings.Contains(s, "context deadline exceeded (Client.Timeout exceeded while awaiting headers)")
            if isDown {
                return model.ErrServerDown
                //replica.Status = model.ServerDown
            }
            //fmt.Println(s)
            return model.ErrFailedNotify

        }
        //client.
        return nil
    }

    return nil
}

func isSuccessNotified(code int) bool {
    return code != http.StatusOK && code != http.StatusNotModified
}

func pingReplica(client *http.Client, replica *model.Server) error {
    if model.ServerDown == replica.Status && time.Since(replica.LastCheck).Seconds() < 2 {
        return ErrServerDown
    }

    replica.LastCheck = time.Now()
    resp, err := http.Get("http://" + replica.Url + "/health")
    if err != nil || resp.StatusCode != http.StatusOK {
        return ErrServerDown
    }
    defer resp.Body.Close()

    return nil
}

type controller struct {
    dataService   *service.DataServiceImple
    isMaster      bool
    replicaList   []model.Server
    instanceIndex int
    notifyChan    *concurrent.List
    cw            *service.ConnectionWatcher
}

func newController(dataService *service.DataServiceImple, isMaster bool, replicaUrls []model.Server, instanceId int, notifyChan *concurrent.List, cw *service.ConnectionWatcher) *controller {
    c := new(controller)
    c.dataService = dataService
    c.instanceIndex = instanceId
    c.isMaster = isMaster
    c.replicaList = replicaUrls
    c.notifyChan = notifyChan
    c.cw = cw
    return c
}

func (c *controller) healthHandler(w http.ResponseWriter, _ *http.Request) {
    w.WriteHeader(http.StatusOK)
}

func (c *controller) getHandler(w http.ResponseWriter, req *http.Request) {
    query := req.URL.Query()
    key := query.Get("key")
    if len(key) == 0 {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    value, code := c.dataService.Get(key)
    if code != http.StatusOK {
        w.WriteHeader(code)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(value))
}

func (c *controller) setHandler(w http.ResponseWriter, req *http.Request) {
    //fmt.Println(req.RemoteAddr, c.cw.Count(), runtime.NumGoroutine())
    //defer func() {
    //    fmt.Println(c.cw.Count(), runtime.NumGoroutine())
    //}()
    //http.MaxBytesReader(w, req.Body, 65535)
    data, err := ioutil.ReadAll(req.Body)
    if err != nil {
        w.WriteHeader(http.StatusRequestEntityTooLarge)
        return
    }

    defer req.Body.Close()
    var item model.Item
    err = json.Unmarshal(data, &item)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    //fmt.Println(item)
    //index := c.getReplicaIndex(c.instanceIndex)
    vectorTime, code := c.dataService.Set(&item, c.instanceIndex, c.getCountInstances())
    if code != http.StatusOK {
        w.WriteHeader(code)
        return
    }

    if c.isMaster {
        for i, replica := range c.replicaList[1:] {
            //if replica.IsMaster {
            //    continue
            //}
            get, ok := c.notifyChan.Get(i)
            if !ok {
                continue
            }
            //serverEvents := get.(*chan model.NotifyServerEvent)
            //fmt.Println(reflect.TypeOf(serverEvents), serverEvents)

            *get <- model.NewNotifyServer(item, replica, i+1, c.instanceIndex, vectorTime)
        }
    }

    w.WriteHeader(http.StatusOK)
}

func (c *controller) getCountInstances() int {
    return len(c.replicaList)
}

func (c *controller) notifyHandler(w http.ResponseWriter, req *http.Request) {
    //fmt.Println(req.RemoteAddr, c.cw.Count(), runtime.NumGoroutine())

    data, err := ioutil.ReadAll(req.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    defer req.Body.Close()
    var event model.NotifyServerEvent
    err = json.Unmarshal(data, &event)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    //fmt.Println(string(data))
    code := c.dataService.SetIfNeeded(&event, c.instanceIndex, c.getCountInstances())
    w.WriteHeader(code)
    if code == http.StatusOK {
        w.Write([]byte("ok"))
    }
    //defer func() {
    //    fmt.Println(c.cw.Count(), runtime.NumGoroutine())
    //}()
}

func (c *controller) getReplicaIndex(instanceId int) int {
    for i, url := range c.replicaList {
        if url.Id == instanceId {
            return i
        }
    }
    return -1
}
