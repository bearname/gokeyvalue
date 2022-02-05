package service

import (
    "fmt"
    "gokeyvalue/pkg/common"
    "gokeyvalue/pkg/model"
    "gokeyvalue/protos"
    "google.golang.org/grpc"
    "time"
)

const connectTimeout = 1 * time.Second
const waitTimeout = 10 * time.Second

type NotifierService interface {
    Publish(item *model.Item, vectorTime []uint64) bool
    OnHandle()
}

type GrpcNotifierService struct {
    isMaster       bool
    replicaList    []model.Server
    selfInstanceId int
    notifyChan     chan model.NotifyServerEvent
    clients        map[int]*protos.KeyValueServiceClient
}

func NewNotifierService(isMaster bool, replicaUrls []model.Server, selfInstanceId int) *GrpcNotifierService {
    s := new(GrpcNotifierService)
    s.isMaster = isMaster
    s.replicaList = replicaUrls
    s.selfInstanceId = selfInstanceId
    s.notifyChan = make(chan model.NotifyServerEvent, 1)
    s.clients = make(map[int]*protos.KeyValueServiceClient)

    return s
}

func (s *GrpcNotifierService) Publish(item *model.Item, vectorTime []uint64) bool {
    if s.isMaster {
        for i, replica := range s.replicaList[1:] {
            if replica.IsMaster {
                continue
            }
            s.notifyChan <- model.NewNotifyServer(*item, replica, i+1, s.selfInstanceId, vectorTime)
        }
        return true
    }
    return false
}

func (s *GrpcNotifierService) OnHandle() {
    for _, server := range s.replicaList {
        if server.Id != s.selfInstanceId {
            conn := common.CreateGrpcConnection(server.Url, common.ConnectTimeout, common.WaitTimeout)
            defer func(conn *grpc.ClientConn) {
               err := conn.Close()
               if err != nil {
                   return
               }
            }(conn)

            client := protos.NewKeyValueServiceClient(conn)
            s.clients[server.Id] = &client
        }
    }

    var client *protos.KeyValueServiceClient
    var ok bool
    for event := range s.notifyChan {
        if event.Replica.IsMaster {
            continue
        }
        client, ok = s.clients[event.Replica.Id]
        if ok{
            ctx, cancel := common.GetCtxWithDeadline(waitTimeout)
            defer cancel()
            _, err := (*client).Notify(ctx, &protos.NotifyEvent{
                Key:         event.Item.Key,
                Value:       event.Item.Value,
                MasterId:    uint64(s.selfInstanceId),
                VectorClock: event.VectorTime,
            })

            if err != nil {
                fmt.Println(err)
            }
        }
    }
}
