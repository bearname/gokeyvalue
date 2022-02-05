package concurrent

import "sync"

type HashMap struct {
    mx   sync.RWMutex
    data map[string]interface{}
}

func NewConcurrentHashMap() *HashMap {
    c := new(HashMap)
    c.data = make(map[string]interface{})
    return c
}
func (s *HashMap) Get(key string) (interface{}, bool) {
    s.mx.RLock()
    defer s.mx.RUnlock()
    i, ok := s.data[key]
    return i, ok
}

func (s *HashMap) Set(key string, value interface{}) {
    s.mx.Lock()
    defer s.mx.Unlock()
    s.data[key] = value
}

func (s *HashMap) Del(key string) bool {
    s.mx.Lock()
    defer s.mx.Unlock()
    _, ok := s.data[key]
    if !ok {
        return false
    }

    delete(s.data, key)
    return true
}
