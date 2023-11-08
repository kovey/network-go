package connection

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/gobwas/ws"
)

type WebSocket struct {
	fd       uint64
	conn     net.Conn
	wQueue   chan []byte
	packet   func(buf []byte) (IPacket, error)
	buf      []byte
	isClosed bool
	lastTime int64
}

func NewWebSocket(fd uint64, conn net.Conn) *WebSocket {
	return &WebSocket{fd, conn, make(chan []byte, CHANNEL_PACKET_MAX), nil, make([]byte, 0, 2097152), false, time.Now().Unix()}
}

func (t *WebSocket) Close() error {
	if t.isClosed {
		return nil
	}

	t.isClosed = true
	return t.conn.Close()
}

func (t *WebSocket) Read(hLen, bLen, bLenOffset int) ([]byte, error) {
	header, err := ws.ReadHeader(t.conn)
	if err != nil {
		return nil, err
	}

	if header.OpCode == ws.OpClose {
		return nil, io.EOF
	}

	hBuf := make([]byte, header.Length)
	_, err = io.ReadFull(t.conn, hBuf)
	if err != nil {
		return nil, err
	}

	if header.Masked {
		ws.Cipher(hBuf, header.Mask, 0)
	}

	return hBuf, nil
}

func (t *WebSocket) WQueue() chan []byte {
	return t.wQueue
}

func (t *WebSocket) FD() uint64 {
	return t.fd
}

func (t *WebSocket) Write(pack []byte) (int, error) {
	if t.isClosed {
		return 0, fmt.Errorf("connection[%d] is closed", t.fd)
	}
	if err := ws.WriteFrame(t.conn, ws.NewBinaryFrame(pack)); err != nil {
		return 0, err
	}

	return len(pack), nil
}

func (t *WebSocket) Send(pack IPacket) error {
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

func (t *WebSocket) Closed() bool {
	return t.isClosed
}

func (t *WebSocket) RemoteIp() string {
	addr := t.conn.RemoteAddr().String()
	return strings.Split(addr, ":")[0]
}

func (t *WebSocket) Expired() bool {
	return time.Now().Unix() > t.lastTime+60
}
