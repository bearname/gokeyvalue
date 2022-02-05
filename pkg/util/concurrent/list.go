package concurrent

import (
    "gokeyvalue/pkg/model"
    "sync"
)

type List struct {
    mx   sync.RWMutex
    data []*chan model.NotifyServerEvent
    size int
}

func NewConcurrentList(countReplica int) *List {
    c := new(List)
    c.data = make([]*chan model.NotifyServerEvent, countReplica)
    c.size = countReplica
    return c
}
func (s *List) Get(key int) (*chan model.NotifyServerEvent, bool) {
    s.mx.RLock()
    defer s.mx.RUnlock()
    if key < 0 || key >= s.size {
        return nil, false
    }
    i := s.data[key]
    return i, true
}

func (s *List) Set(key int, value *chan model.NotifyServerEvent) bool {
    if key < 0 || key >= s.size {
        return false
    }
    s.mx.Lock()
    defer s.mx.Unlock()

    s.data[key] = value
    return false
}
//
//func (s *ConcurrentHashMap) Del(key int) bool {
//    s.mx.RLock()
//    defer s.mx.RUnlock()
//    _, ok := s.data[key]
//    if !ok {
//        return false
//    }
//
//    delete(s.data, key)
//    return true
//}
