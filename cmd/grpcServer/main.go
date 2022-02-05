package main

import (
    "errors"
    "flag"
    "gokeyvalue/pkg/model"
    "gokeyvalue/pkg/repo"
    "gokeyvalue/pkg/service"
    transpGrpc "gokeyvalue/pkg/transport/grpc"
    "gokeyvalue/protos"
    "google.golang.org/grpc"
    "log"
    "net"
    "net/http"
    "net/http/pprof"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "sync"
    "syscall"

    //"github.com/golang/protobuf/proto"
    "time"
)

func main() {
    go func() {
        r := http.NewServeMux()

        // Регистрация pprof-обработчиков
        r.HandleFunc("/debug/pprof/", pprof.Index)
        r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
        r.HandleFunc("/debug/pprof/profile", pprof.Profile)
        r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
        r.HandleFunc("/debug/pprof/trace", pprof.Trace)

        http.ListenAndServe(":8080", r)
    }()
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

    //parseGuarantees, err := model.ParseGuarantees(guarantees)
    //if err != nil {
    //    log.Fatal(err)
    //}

    urls := strings.Split(otherServer, ",")

    if len(urls) == 0 {
        isMaster = true
    }
    var replicaList []model.Server
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
        }
    }

    //err := os.Mkdir(volume, 0666)
    //pwd, _ := os.Getwd()
    //root := pwd + string(os.PathSeparator) + volume +string(os.PathSeparator)
    //name := root  +  "db.bin"
    //
    //
    //dataFile, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND, 0666)
    //if err != nil {
    //    log.Fatal(err)
    //}
    //defer dataFile.Close()
    //hashFile, err := os.OpenFile(root+"hash.txt", os.O_CREATE|os.O_RDWR, 0666)
    //if err != nil {
    //    log.Fatal(err)
    //}
    //defer dataFile.Close()

    //repository := repo.NewFileRepo(dataFile, hashFile)
    repository := repo.NewMemoryRepo()
    dataService := service.NewDataService(repository)

    err := dataService.RestoreHashMap()
    if err != nil {
        log.Fatal(err)
    }
    go func() {
        for {
            dataService.BackupHashMap()
            time.Sleep(10 * time.Second)
        }
    }()

    notifierService := service.NewNotifierService(isMaster, replicaList, instanceId)

    wg := sync.WaitGroup{}
    wg.Add(1)
    //countNotifier := len(replicaList)

    //-instanceId=1 -port=8002 -isMaster=false -urls=1!:8001 -volume=inst-1\
    //-instanceId=0 -port=8000 -isMaster=true -urls=0!:8000,1!:8001 -volume=inst-0\
    srv := transpGrpc.NewGrpcServer(dataService, notifierService, len(replicaList), instanceId)
    s := grpc.NewServer()
    defer s.GracefulStop()
    protos.RegisterKeyValueServiceServer(s, srv)

    log.Println("Started at " + time.Now().String())
    l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
    if err != nil {
        log.Fatal(err)
    }

    //proto.Marshal()
    if isMaster {
        go func() {
            defer wg.Done()
            for i := 0; i < 10; i++ {
                wg.Add(1)
                go notifierService.OnHandle()
            }
        }()
    }

    if err = s.Serve(l); err != nil {
        log.Fatal(err)
    }

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
}
