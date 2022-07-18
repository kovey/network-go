package websocket

import (
	"fmt"
	"github.com/kovey/network-go/connection"

	"golang.org/x/net/websocket"
)

type WebSocket struct {
	fd       int
	conn     *websocket.Conn
	rQueue   chan connection.IPacket
	wQueue   chan connection.IPacket
	packet   func(buf []byte) (connection.IPacket, error)
	buf      []byte
	isClosed bool
}

func Dial(protocol, host string, port int, path string) (*websocket.Conn, error) {
	return websocket.Dial(fmt.Sprintf("%s://%s:%d/%s", protocol, host, port, path), "", fmt.Sprintf("%s://%s", "http", host))
}

func NewWebSocket(fd int, conn *websocket.Conn) *WebSocket {
	return &WebSocket{
		fd, conn, make(chan connection.IPacket, connection.CHANNEL_PACKET_MAX),
		make(chan connection.IPacket, connection.CHANNEL_PACKET_MAX), nil, make([]byte, 0, 2097152), false,
	}
}

func (t *WebSocket) Close() error {
	if t.isClosed {
		return nil
	}

	close(t.rQueue)
	close(t.wQueue)
	t.isClosed = true
	return t.conn.Close()
}

func (t *WebSocket) Read(hLen, bLen, bLenOffset int) ([]byte, error) {
	var hBuf = make([]byte, 2097152)
	n, err := t.conn.Read(hBuf)
	if err != nil {
		return nil, err
	}

	return hBuf[:n], nil
}

func (t *WebSocket) WQueue() chan connection.IPacket {
	return t.wQueue
}

func (t *WebSocket) RQueue() chan connection.IPacket {
	return t.rQueue
}

func (t *WebSocket) FD() int {
	return t.fd
}

func (t *WebSocket) Write(pack connection.IPacket) (int, error) {
	if t.isClosed {
		return 0, fmt.Errorf("connection[%d] is closed", t.fd)
	}
	return t.conn.Write(pack.Serialize())
}

func (t *WebSocket) Closed() bool {
	return t.isClosed
}
