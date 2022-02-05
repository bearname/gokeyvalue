package service

import (
    "net"
    "net/http"
    "sync/atomic"
)

type ConnectionWatcher struct {
    n int64
}

func (cw *ConnectionWatcher) OnStateChange(conn net.Conn, state http.ConnState) {
    //fmt.Println(state)
    switch state {
    case http.StateNew:
        atomic.AddInt64(&cw.n, 1)
    case http.StateHijacked, http.StateClosed:
        atomic.AddInt64(&cw.n, -1)
    }
}

func (cw *ConnectionWatcher) Count() int {
    return int(atomic.LoadInt64(&cw.n))
}

