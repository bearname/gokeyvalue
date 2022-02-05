package grpc

import (
    "context"
    "errors"
    "gokeyvalue/pkg/model"
    "gokeyvalue/pkg/service"
    "gokeyvalue/protos"
    "net/http"
    "strconv"
)

type Server struct {
    dataService     *service.DataServiceImple
    notifierService service.NotifierService
    isMaster        bool
    countReplica int
    selfIndex    int
}

func NewGrpcServer(dataService *service.DataServiceImple, notifierService service.NotifierService, countReplica, selfIndex int) *Server {
    s := new(Server)
    s.dataService = dataService
    s.notifierService = notifierService
    s.countReplica = countReplica
    s.selfIndex = selfIndex
    return s
}

func (s *Server) Set(ctx context.Context, item *protos.Item) (*protos.Success, error) {
    it := model.Item{Key: item.Key, Value: item.Value}
    vectorTime, code := s.dataService.Set(&it, s.selfIndex, s.countReplica)
    if code != http.StatusOK {
        return &protos.Success{Success: false}, nil
    }

    success := s.notifierService.Publish(&it, vectorTime)

    return &protos.Success{Success: success}, nil
}

func (s *Server) Notify(ctx context.Context, event *protos.NotifyEvent) (*protos.Success, error) {
    //fmt.Println(event.Key, event.Value, event.MasterId, event.VectorClock)
    serverEvent := model.NotifyServerEvent{
        Item: model.Item{
            Key: event.Key, Value: event.Value,
        },
        ReplicaIndex: s.selfIndex,
        InstanceId: int(event.MasterId),
        VectorTime:   event.VectorClock,
    }

    //_, code := s.dataService.Set(&serverEvent, int(event.MasterId), 2)
    code := s.dataService.SetIfNeeded(&serverEvent, s.selfIndex, s.countReplica)

    if code == http.StatusOK {
        return &protos.Success{
            Success: true,
        }, nil
    }
    return &protos.Success{
        Success: false,
    }, nil
}

func (s *Server) Get(ctx context.Context, key *protos.GetKey) (*protos.ItemValue, error) {
    //fmt.Println(&key.Key)
    value, err:= s.dataService.Get(key.Key)
    if err != http.StatusOK {
        return nil, errors.New(strconv.Itoa(err))
    }
    return &protos.ItemValue{
        Value: value,
    }, nil
}

func (s *Server) DeleteKey(context.Context, *protos.GetKey) (*protos.Success, error) {
    return &protos.Success{
        Success: true,
    }, nil
}
