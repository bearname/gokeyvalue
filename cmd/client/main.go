package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "gokeyvalue/pkg/model"
    "gokeyvalue/pkg/transport"
    "io/ioutil"
    "math/rand"
    "net/http"
    "runtime"
    "strconv"
    "sync"
    "time"
)

var rep []string
var mx sync.RWMutex

func main() {
    mx = sync.RWMutex{}
    rep = append(rep, "sdf")
    var count int
    var isRand bool
    var isGet bool
    var requestParallel int
    var port int
    flag.IntVar(&port, "port", 8000, "port")
    flag.IntVar(&count, "c", 100, "count request to key valueserver")
    flag.IntVar(&requestParallel, "req", 10, "parallel request")
    flag.BoolVar(&isRand, "r", false, "random key")
    flag.BoolVar(&isGet, "isGet", true, "random key")
    flag.Parse()
    wg := sync.WaitGroup{}
    rand.Seed(time.Now().UnixNano())
    start := time.Now()
    maxGor := 0
    goroutine := runtime.NumGoroutine()
    if goroutine > maxGor {
        maxGor = goroutine
    }
    countSuccess := 0
    fmt.Println("init", goroutine)
    ch := make(chan int)
    wg.Add(1)
    go func() {
        defer wg.Done()
        for i := 0; i < requestParallel; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                client := transport.HttpClient(20)

                for key := range ch {
                    //now := time.Now()
                    baseUrl := "http://localhost:" + strconv.Itoa(port)
                    if isGet {
                        resp, err := http.Get(baseUrl + "?key=" + strconv.Itoa(key))
                        handleResp(err, resp)
                        if err == nil && resp.StatusCode == http.StatusOK {
                            countSuccess++
                        }
                    } else {
                        item := model.Item{
                            Key:   strconv.Itoa(key),
                            Value: strconv.Itoa(key),
                        }
                        data, err := json.Marshal(item)
                        if err != nil {
                            return
                        }
                        buffer := bytes.NewBuffer(data)
                        code, err := transport.SendRequest(client, http.MethodPost, baseUrl, buffer)
                        //resp, err := http.Post(baseUrl, "application/json; charset: utf-8;", bytes.NewBuffer(data))
                        //handleResp(err, resp)
                        //code = resp.StatusCode
                        if err == nil && (code == http.StatusOK || code == http.StatusNotModified) {
                            countSuccess++
                        }
                    }

                    goroutine = runtime.NumGoroutine()
                    if goroutine > maxGor {
                        maxGor = goroutine
                    }
                    //fmt.Println("goroutine", goroutine)
                    //fmt.Println(time.Since(now).Milliseconds(), "milli sec")
                }
            }()
        }
    }()
    for i := 0; i < count; i++ {
        var key int64
        if isRand {
            key = int64(rand.Intn(1000000))
        } else {
            key = int64(i)
        }
        fmt.Println(key)
        ch <- int(key)
    }
    close(ch)

    wg.Wait()
    end := time.Since(start)
    fmt.Println(rep)
    fmt.Println("complete", end.Milliseconds(), "milli sec", "max goroutine", maxGor)
    fmt.Println(end.Seconds(), "s")
    fmt.Println(end.String())
    fmt.Println("success", countSuccess)
    fmt.Println("failed", count-countSuccess)
    seconds := int(end.Seconds())
    if seconds == 0 {
        seconds = 1
    }
    fmt.Println(countSuccess / seconds)
}

func handleResp(err error, resp *http.Response) {
    if err != nil {
        fmt.Println(err)
    } else {
        //fmt.Println("resp.StatusCode", resp.StatusCode)
        data, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Println(err)
        } else {
            fmt.Println(string(data))
        }
    }
}

//
//client 50
//complete 120317 milli sec max goroutine 7071
//120.3170327 s
//2m0.3170327s
//success 100000
//failed 0
//833
//
//
//client 10
//complete 118437 milli sec max goroutine 6187
//118.4374108 s
//1m58.4374108s
//success 100000
//failed 0
//847
