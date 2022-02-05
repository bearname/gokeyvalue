package util

import (
    "gokeyvalue/pkg/model"
    "sync"
)

type HashMap struct {
    data map[string]model.SavedInFileValue
    mx   sync.RWMutex
}

func NewHashMap() *HashMap {
    h := new(HashMap)
    h.data = make(map[string]model.SavedInFileValue)
    return h
}

func NewSavedInFileValue(seek, length int64, instanceId int, vectorTime []uint64) *model.SavedInFileValue {
    s := new(model.SavedInFileValue)
    s.Seek = seek
    s.Length = length
    s.InstanceId = instanceId
    s.VectorTime = vectorTime
    return s
}

func (s *HashMap) Get(key string) (*model.SavedInFileValue, bool) {
    s.mx.RLock()
    value, ok := s.data[key]
    s.mx.RUnlock()
    return &value, ok
}

func (s *HashMap) Set(key string, value *model.SavedInFileValue) {
    s.mx.Lock()
    s.data[key] = *value
    s.mx.Unlock()
}

func (s *HashMap) List() ([]string, []model.SavedInFileValue) {
    keys := make([]string, 0, len(s.data))
    values := make([]model.SavedInFileValue, 0, len(s.data))
    s.mx.RLock()

    for k, v := range s.data {
        keys = append(keys, k)
        values = append(values, v)
    }
    s.mx.RUnlock()
    return keys, values
}
