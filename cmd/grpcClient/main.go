package main

import (
    "context"
    "flag"
    "fmt"
    log "github.com/sirupsen/logrus"
    "gokeyvalue/pkg/common"
    "gokeyvalue/protos"
    "google.golang.org/grpc"
    "strconv"
    "sync"
    "time"
)

const connectTimeout = 1 * time.Second
const waitTimeout = 1 * time.Second

func main() {
    //flag.Parse()
    //if flag.NArg() < 2 {
    //    log.Fatal("not enough arguments")
    //}

    //x, err := strconv.Atoi(flag.Arg(0))
    //if err != nil {
    //    log.Fatal(err)
    //}
    //y, err := strconv.Atoi(flag.Arg(1))
    //if err != nil {
    //    log.Fatal(err)
    //}
    var address int
    var key string
    flag.IntVar(&address, "p", 8000, "port")
    flag.StringVar(&key, "key", "", "port")
    flag.Parse()
    if len(key) == 0 {
        fmt.Println("key length  must be > 0")
        return
    }

    conn := common.CreateGrpcConnection(":"+strconv.Itoa(address), connectTimeout, waitTimeout)
    defer func(conn *grpc.ClientConn) {
        err := conn.Close()
        if err != nil {
            log.Fatal(err)
        }
    }(conn)

    client := protos.NewKeyValueServiceClient(conn)
    var item []uint64
    item = append(item, 1, 2, 3)
    wg := sync.WaitGroup{}
    wg.Add(1)
    ch := make(chan int)
    count := 100000
    go func() {
        defer wg.Done()
        for i := 0; i < count; i++ {
            ch <- i
        }
        close(ch)
    }()

    now := time.Now()
    req, err := getReq(client, key)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(req)
    end := time.Since(now)
    i := int(end.Seconds())
    if i == 0 {
        i = 1
    }

    fmt.Println(end.Nanoseconds(), "nanosec req/sec", count/i)
    fmt.Println(end.Milliseconds(), "micro req/sec", count/i)
    fmt.Println(end.Microseconds(), "micro req/sec", count/i)
    fmt.Println(end.Seconds(), "sec req/sec", count/i)

    //listResult(client, request)
    //
    //ctx, cancel := getCtxWithTimeout()
    //if ctx.Err() == context.Canceled {
    //	log.Println(codes.Canceled, "Client cancelled, abandoning.")
    //}
    //defer cancel()
    //stream, err := client.RecordResult(ctx)
    //if err != nil {
    //	log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
    //}
    //var points []*protos.AddRequest
    //for i := 0; i < 5; i++ {
    //	points = append(points, randomPoint())
    //}
    //
    //fmt.Println(points)
    //
    //for _, point := range points {
    //	if err = stream.Send(point); err != nil {
    //		log.Fatalf("%v.Send(%v) = %v", stream, point, err)
    //	}
    //}
    //reply, err := stream.CloseAndRecv()
    //if err != nil {
    //	log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
    //}
    //log.Printf("Route summary: %v", reply)
}

//
//func randomPoint() *protos.AddRequest {
//    return &protos.AddRequest{
//        X: int32(rand.Intn(5)),
//        Y: int32(rand.Intn(5)),
//    }
//}

func notifyReq(item int, client protos.KeyValueServiceClient, event *protos.NotifyEvent) {
    ctx, cancel := getCtxWithTimeout()
    defer cancel()
    _, err := client.Notify(ctx, event)
    //key, err := client.Get(ctx, &protos.GetKey{
    //    Key: event.Key,
    //})

    if err != nil {
        log.Fatal(err)
    }
    //fmt.Println(key.Key)

    //fmt.Println(addResponse.Success)
}

func getReq(client protos.KeyValueServiceClient, key string) (string, error) {
    ctx, cancel := getCtxWithTimeout()
    defer cancel()
    value, err := client.Get(ctx, &protos.GetKey{Key: key})
    if err != nil {
        return "", err
    }
    return value.Value, nil
}

//
//func listResult(client protos.AdderClient, request *protos.AddRequest) {
//    ctx, cancel := getCtxWithTimeout()
//    defer cancel()
//    streamResult, err := client.ListResult(ctx, request)
//    if err != nil {
//        log.Fatal(err)
//    }
//
//    streamResult.Trailer()
//
//    for {
//        result, err := streamResult.Recv()
//        if err == io.EOF {
//            break
//        }
//        if err != nil {
//            log.Fatalf("%v.ListResult(_) = _, %v", client, err)
//        }
//
//        fmt.Printf("Result = %d", result.GetResult())
//    }
//}

func getCtxWithTimeout() (context.Context, context.CancelFunc) {
    ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(waitTimeout))
    return ctx, cancel
}
