package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/kovey/network-go/connection"

	"github.com/gobwas/ws"
	"github.com/kovey/logger-go/logger"
)

type WebSocketService struct {
	connMax   int
	connCount int
	curFD     int
	listener  net.Listener
	locker    sync.Mutex
	isClosed  bool
}

func NewWebSocketService(connMax int) *WebSocketService {
	return &WebSocketService{connMax, 0, 0, nil, sync.Mutex{}, false}
}

func (t *WebSocketService) IsClosed() bool {
	return t.isClosed
}

func (t *WebSocketService) Listen(host string, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	logger.Debug("server listen on %s:%d", host, port)

	t.listener = listener
	return nil
}

func (t *WebSocketService) Accept() (connection.IConnection, error) {
	if t.connCount > t.connMax {
		return nil, fmt.Errorf("connection is reach max[%d]", t.connMax)
	}

	conn, err := t.listener.Accept()
	if err != nil {
		return nil, err
	}

	_, err = ws.Upgrade(conn)
	if err != nil {
		return nil, err
	}

	t.connCount++
	t.curFD++
	return connection.NewWebSocket(t.curFD, conn), nil
}

func (t *WebSocketService) Close() {
	t.locker.Lock()
	t.connCount--
	t.locker.Unlock()
}

func (t *WebSocketService) Shutdown() {
	t.isClosed = true
	t.listener.Close()
}
