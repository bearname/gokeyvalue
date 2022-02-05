package service

import (
    "errors"
    "gokeyvalue/pkg/transport"
    "net/http"
    "sync"
)

var ErrNotFound = errors.New("not found")

type ConnectorService struct {
    mx      sync.Mutex
    clients map[string]*http.Client
    size    int
}

func NewConnectorService(size int) *ConnectorService {
    c := new(ConnectorService)
    c.clients = make(map[string]*http.Client, size)
    c.size = size
    return c
}

func (s *ConnectorService) AddOrUpdateClient(instanceId string, client *http.Client) *http.Client {
    s.mx.Lock()
    defer s.mx.Unlock()
    _, ok := s.clients[instanceId]
    if !ok {
        s.size++
    }
    s.clients[instanceId] = client
    return client
}

func (s *ConnectorService) AddOrUpdateClientDefault(instanceId string) *http.Client {
    s.mx.Lock()
    defer s.mx.Unlock()
    _, ok := s.clients[instanceId]
    if !ok {
        s.size++
    }
    client := transport.HttpClient(20)
    s.clients[instanceId] = client
    return client
}

func (s *ConnectorService) GetClient(instanceId string) (*http.Client, error) {
    s.mx.Lock()
    defer s.mx.Unlock()
    client, ok := s.clients[instanceId]
    if !ok {
        return nil, ErrNotFound
    }
    return client, nil
}

func (s *ConnectorService) DeleteClient(instanceId string) error {
    s.mx.Lock()
    defer s.mx.Unlock()
    _, ok := s.clients[instanceId]
    if !ok {
        return ErrNotFound
    }

    s.size--
    delete(s.clients, instanceId)
    return nil
}

func (s *ConnectorService) Keys() []string {
    s.mx.Lock()
    defer s.mx.Unlock()
    var keys []string
    for key := range s.clients {
        keys = append(keys, key)
    }

    return keys
}

func (s *ConnectorService) GetSize() int {
    s.mx.Lock()
    defer s.mx.Unlock()
    return s.size
}
