package connection

import (
	"fmt"
	"io"
	"net"

	"github.com/gobwas/ws"
)

type WebSocket struct {
	fd       int
	conn     net.Conn
	rQueue   chan IPacket
	wQueue   chan []byte
	packet   func(buf []byte) (IPacket, error)
	buf      []byte
	isClosed bool
}

func NewWebSocket(fd int, conn net.Conn) *WebSocket {
	return &WebSocket{fd, conn, make(chan IPacket, CHANNEL_PACKET_MAX), make(chan []byte, CHANNEL_PACKET_MAX), nil, make([]byte, 0, 2097152), false}
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

	header.Masked = false
	ws.WriteHeader(t.conn, header)

	return hBuf, nil
}

func (t *WebSocket) WQueue() chan []byte {
	return t.wQueue
}

func (t *WebSocket) RQueue() chan IPacket {
	return t.rQueue
}

func (t *WebSocket) FD() int {
	return t.fd
}

func (t *WebSocket) Write(pack []byte) (int, error) {
	if t.isClosed {
		return 0, fmt.Errorf("connection[%d] is closed", t.fd)
	}
	return t.conn.Write(pack)
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
