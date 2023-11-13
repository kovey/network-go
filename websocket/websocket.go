package websocket

import (
	"fmt"
	"strings"

	"github.com/kovey/network-go/connection"

	"golang.org/x/net/websocket"
)

type WebSocket struct {
	fd       uint64
	conn     *websocket.Conn
	wQueue   chan []byte
	sQueue   chan bool
	buf      []byte
	isClosed bool
	ext      map[int64]any
}

func Dial(protocol, host string, port int, path string) (*websocket.Conn, error) {
	return websocket.Dial(fmt.Sprintf("%s://%s:%d/%s", protocol, host, port, path), "", fmt.Sprintf("%s://%s", "http", host))
}

func NewWebSocket(fd uint64, conn *websocket.Conn) *WebSocket {
	return &WebSocket{
		fd: fd, conn: conn, wQueue: make(chan []byte, connection.CHANNEL_PACKET_MAX),
		buf: make([]byte, 0, 2097152), isClosed: false, ext: make(map[int64]any), sQueue: make(chan bool, 5),
	}
}

func (t *WebSocket) Close() error {
	if t.isClosed {
		return nil
	}

	t.isClosed = true
	t.sQueue <- true
	return t.conn.Close()
}

func (t *WebSocket) Read(hLen, bLen, bLenOffset int) ([]byte, error) {
	var hBuf = make([]byte, 2097152)
	n, err := t.conn.Read(hBuf)
	fmt.Println("n: ", n, "buf:", hBuf[:n])
	if err != nil {
		return nil, err
	}

	return hBuf[:n], nil
}

func (t *WebSocket) WQueue() <-chan []byte {
	return t.wQueue
}

func (t *WebSocket) FD() uint64 {
	return t.fd
}

func (t *WebSocket) Write(pack []byte) (int, error) {
	if t.isClosed {
		return 0, fmt.Errorf("connection[%d] is closed", t.fd)
	}
	return t.conn.Write(pack)
}

func (t *WebSocket) Closed() bool {
	return t.isClosed
}

func (t *WebSocket) Send(pack connection.IPacket) error {
	buf := pack.Serialize()
	if buf == nil {
		return fmt.Errorf("pack is empty")
	}

	return t.SendBytes(buf)
}

func (t *WebSocket) SendBytes(buf []byte) error {
	if t.isClosed {
		return fmt.Errorf("connection[%d] is closed", t.fd)
	}

	t.wQueue <- buf
	return nil
}

func (t *WebSocket) RemoteIp() string {
	addr := t.conn.RemoteAddr().String()
	return strings.Split(addr, ":")[0]
}

func (t *WebSocket) Expired() bool {
	return false
}

func (t *WebSocket) Set(key int64, val any) {
	t.ext[key] = val
}

func (t *WebSocket) Get(key int64) (any, bool) {
	val, ok := t.ext[key]
	return val, ok
}

func (t *WebSocket) SQueue() <-chan bool {
	return t.sQueue
}
