package main

import (
    "context"
    "fmt"
    "gokeyvalue/pkg/common"
    "gokeyvalue/protos"
    "google.golang.org/grpc"
    "log"
    "strconv"
    "testing"
    "time"
)

func BenchmarkGrpcSetTest(b *testing.B) {
    address := strconv.Itoa(8000)
    conn := common.CreateGrpcConnection(":" + address, common.ConnectTimeout, common.WaitTimeout)

    defer func(conn *grpc.ClientConn) {
        err := conn.Close()
        if err != nil {
            fmt.Println(err)
        }
    }(conn)

    var request protos.Item
    client := protos.NewKeyValueServiceClient(conn)
    var key string
    var value string
    var err error
    var ctx context.Context
    var cancel context.CancelFunc

    for it := 0; it < b.N; it++ {
        key = "hello" + strconv.Itoa(it)
        //fmt.Println(key)
        value = "world" + strconv.Itoa(it)
        request = protos.Item{Key: key, Value: value}
        //request = protos.NotifyEvent{Key: key, Value: value, MasterId: 1, VectorClock: item}
        //err := notifyReq(it, &client, &request)
        ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(common.WaitTimeout))
        defer cancel()
        _, err = client.Set(ctx, &request)
        //err = setReq(&client, &request)
        if err != nil {
            fmt.Println(err)
            time.Sleep(1 * time.Second)
            conn = common.CreateGrpcConnection(":" + address, common.ConnectTimeout, common.WaitTimeout)
            defer func(conn *grpc.ClientConn) {
                err = conn.Close()
                if err != nil {
                    log.Fatal(err)
                }
            }(conn)

            client = protos.NewKeyValueServiceClient(conn)
        }
    }
}

func GetCtxWithTimeout() (context.Context, context.CancelFunc) {
    ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(common.WaitTimeout))
    return ctx, cancel
}