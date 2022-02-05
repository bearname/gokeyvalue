package model

import (
    "errors"
    "fmt"
)

type ReplicationGuarantees int

const (
    GuaranteesAsync ReplicationGuarantees = iota
    GuaranteesSync
)

var ErrUnsupportedGuarantee = errors.New("unsupported guarantee")

func ParseGuarantees(value int) (ReplicationGuarantees, error) {
    switch value {
    case int(GuaranteesAsync):
        return GuaranteesAsync, nil
    case int(GuaranteesSync):
        return GuaranteesSync, nil
    default:
        return 0, ErrUnsupportedGuarantee
    }
}
func (e ReplicationGuarantees) String() string {
    switch e {
    case GuaranteesAsync:
        return "async"
    case GuaranteesSync:
        return "sync"
    default:
        return fmt.Sprintf("%d", int(e))
    }
}

type Item struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

type NotifyServerEvent struct {
    Replica      Server `json:"replica"`
    ReplicaIndex int    `json:"replica_index"`
    Item         Item   `json:"item"`
    InstanceId   int    `json:"instance_id"`
    VectorTime   []uint64  `json:"vector_time"`
}

func NewNotifyServer(item Item, replica Server, instanceId int, replicaIndex int, vectorTime []uint64) NotifyServerEvent {
    n := NotifyServerEvent{}
    n.Item = item
    n.Replica = replica
    n.ReplicaIndex = replicaIndex
    n.InstanceId = instanceId
    n.VectorTime = vectorTime
    return n
}

type ServerStatus int

const (
    ServerOk ServerStatus = iota
    ServerDown
)

type Server struct {
    Id        int          `json:"id"`
    Url       string       `json:"url"`
    IsMaster  bool         `json:"is_master"`
    Status    ServerStatus `json:"status"`
}

func NewServer(id int, url string, isMaster bool) *Server {
    s := new(Server)
    s.Id = id
    s.Url = url
    s.Status = ServerOk
    s.IsMaster = isMaster
    return s
}

type Value struct {
    InstanceId int      `json:"instance_id"`
    VectorTime []uint64 `json:"vector_time"`
}

type MemoryValue struct {
    Value
    Data string `json:"data"`
}

type SavedInFileValue struct {
    Value
    Seek   int64 `json:"s"`
    Length int64 `json:"l"`
}
