package server

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/network-go/v2/connection"
)

type TcpService struct {
	connMax       int
	connCount     int
	curFD         uint64
	listener      net.Listener
	locker        sync.Mutex
	isClosed      bool
	maxLen        int
	bodyLengthLen int
	bodyLenOffset int
	headerLenType connection.HeaderLenType
	endian        binary.ByteOrder
}

func NewTcpService(connMax int) *TcpService {
	return &TcpService{connMax: connMax, locker: sync.Mutex{}}
}

func (c *TcpService) HeaderLenType(t connection.HeaderLenType) *TcpService {
	c.headerLenType = t
	return c
}

func (c *TcpService) Endian(e binary.ByteOrder) *TcpService {
	c.endian = e
	return c
}

func (c *TcpService) MaxLen(maxLen int) *TcpService {
	c.maxLen = maxLen
	return c
}

func (c *TcpService) BodyLenghLen(length int) *TcpService {
	c.bodyLengthLen = length
	return c
}

func (c *TcpService) BodyLenOffset(offset int) *TcpService {
	c.bodyLenOffset = offset
	return c
}

func (t *TcpService) IsClosed() bool {
	return t.isClosed
}

func (t *TcpService) Listen(host string, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	debug.Info("server listen on %s:%d", host, port)

	t.listener = listener
	return nil
}

func (t *TcpService) Accept() (*connection.Connection, error) {
	if t.connCount > t.connMax {
		return nil, fmt.Errorf("connection is reach max[%d]", t.connMax)
	}

	conn, err := t.listener.Accept()
	if err != nil {
		return nil, err
	}

	t.connCount++
	t.curFD++
	return connection.NewConnection(t.curFD, conn).HeaderLenType(t.headerLenType).Endian(t.endian).MaxLen(t.maxLen).BodyLenghLen(t.bodyLengthLen).BodyLenOffset(t.bodyLenOffset), nil
}

func (t *TcpService) Close() {
	t.locker.Lock()
	t.connCount--
	t.locker.Unlock()
}

func (t *TcpService) Shutdown() {
	t.isClosed = true
	t.listener.Close()
}
