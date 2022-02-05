package main

import (
    "context"
    "flag"
    "fmt"
    log "github.com/sirupsen/logrus"
    "gokeyvalue/protos"
    "google.golang.org/grpc"
    "google.golang.org/grpc/connectivity"
    "google.golang.org/grpc/credentials/insecure"
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
    var countReq int
    flag.IntVar(&address, "p", 8000, "port")
    flag.IntVar(&countReq, "c", 1000, "port")
    flag.Parse()

    var item []uint64
    item = append(item, 1, 2, 3)
    wg := sync.WaitGroup{}
    wg.Add(1)
    ch := make(chan int, 1)
    go func() {
        defer wg.Done()
        for i := 0; i < countReq; i++ {
            ch <- i
        }
        close(ch)
    }()

    now := time.Now()
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            conn := createGrpcConnection(":" + strconv.Itoa(address))
            defer func(conn *grpc.ClientConn) {
                err := conn.Close()
                if err != nil {
                    log.Fatal(err)
                }
            }(conn)

            client := protos.NewKeyValueServiceClient(conn)

            defer wg.Done()
            var request protos.Item
            for it := range ch {
                key := "hello" + strconv.Itoa(it)
                //fmt.Println(key)
                value := "world" + strconv.Itoa(it)
                request = protos.Item{Key: key, Value: value}
                //request = protos.NotifyEvent{Key: key, Value: value, MasterId: 1, VectorClock: item}
                //err := notifyReq(it, &client, &request)
                err := SetReq(&client, &request)
                if err != nil {
                    fmt.Println(err)
                    time.Sleep(1 * time.Second)
                    conn = createGrpcConnection(":" + strconv.Itoa(address))
                    defer func(conn *grpc.ClientConn) {
                        err := conn.Close()
                        if err != nil {
                            log.Fatal(err)
                        }
                    }(conn)

                    client = protos.NewKeyValueServiceClient(conn)
                }
                //fmt.Println(it)
            }
        }()
    }

    wg.Wait()
    end := time.Since(now)
    i := int(end.Seconds())
    if i == 0 {
        i = 1
    }
    fmt.Println(end.String(), "count=", countReq, "req/sec", countReq/i)

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

func notifyReq(item int, client *protos.KeyValueServiceClient, event *protos.NotifyEvent) error {
    ctx, cancel := getCtxWithTimeout()
    defer cancel()
    _, err := (*client).Notify(ctx, event)
    //key, err := client.Get(ctx, &protos.GetKey{
    //    Key: event.Key,
    //})

    if err != nil {
        log.Println(err)
        return err
    }
    return nil
    //fmt.Println(key.Key)

    //fmt.Println(addResponse.Success)
}
func SetReq( client *protos.KeyValueServiceClient, item *protos.Item) error {
    ctx, cancel := getCtxWithTimeout()
    defer cancel()
    _, err := (*client).Set(ctx, item)
    //key, err := client.Get(ctx, &protos.GetKey{
    //    Key: event.Key,
    //})

    if err != nil {
        log.Println(err)
        return err
    }
    return nil
    //fmt.Println(key.Key)

    //fmt.Println(addResponse.Success)
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

func createGrpcConnection(address string) *grpc.ClientConn {
    conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }

    timer := time.After(connectTimeout)
    for conn.GetState() != connectivity.Ready {
        select {
        case <-timer:
            log.WithField("url", address).Fatal("GRPC connection timeout")
        default:
        }

        ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(waitTimeout))
        conn.WaitForStateChange(ctx, conn.GetState())
        cancel()
    }
    return conn
}
