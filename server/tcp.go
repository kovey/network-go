package server

import (
	"fmt"
	"net"
	"network/connection"
	"sync"

	"github.com/kovey/logger-go/logger"
)

type TcpService struct {
	connMax   int
	connCount int
	curFD     int
	listener  net.Listener
	locker    sync.Mutex
}

func NewTcpService(connMax int) *TcpService {
	return &TcpService{connMax, 0, 0, nil, sync.Mutex{}}
}

func (t *TcpService) Listen(host string, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	logger.Debug("server listen on %s:%d", host, port)

	t.listener = listener
	return nil
}

func (t *TcpService) Accept() (connection.IConnection, error) {
	if t.connCount > t.connMax {
		return nil, fmt.Errorf("connection is reach max[%d]", t.connMax)
	}

	conn, err := t.listener.Accept()
	if err != nil {
		return nil, err
	}

	t.connCount++
	t.curFD++
	return connection.NewTcp(t.curFD, conn), nil
}

func (t *TcpService) Close() {
	t.locker.Lock()
	t.connCount--
	t.locker.Unlock()
}

func (t *TcpService) Shutdown() {
	t.listener.Close()
}
