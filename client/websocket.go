package client

import (
	"github.com/kovey/network-go/connection"
	"github.com/kovey/network-go/websocket"
)

type WebSocket struct {
	conn *websocket.WebSocket
}

func NewWebSocket() *WebSocket {
	return &WebSocket{}
}

func (t *WebSocket) Dial(host string, port int) error {
	conn, err := websocket.Dial("ws", host, port, "/")
	if err != nil {
		return err
	}

	t.conn = websocket.NewWebSocket(1, conn)
	return nil
}

func (t *WebSocket) Connection() connection.IConnection {
	return t.conn
}
