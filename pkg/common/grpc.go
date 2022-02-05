package common

import (
    "context"
    log "github.com/sirupsen/logrus"
    "google.golang.org/grpc"
    "google.golang.org/grpc/connectivity"
    "google.golang.org/grpc/credentials/insecure"
    "time"
)

const ConnectTimeout = 1 * time.Second
const WaitTimeout = 1 * time.Second

func CreateGrpcConnection(address string, connectTimeout, waitTimeout time.Duration) *grpc.ClientConn {
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

func GetCtxWithDeadline(waitTimeout time.Duration) (context.Context, context.CancelFunc) {
    ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(waitTimeout))
    return ctx, cancel
}
