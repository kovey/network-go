package websocket

import (
	"fmt"
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/network-go/client"
	"github.com/kovey/network-go/connection"
	"github.com/kovey/network-go/server"
)

type Handler struct {
}

func (h *Handler) Connect(conn connection.IConnection) error {
	fmt.Printf("new connection[%d]\n", conn.FD())
	return nil
}

func (h *Handler) Receive(context *server.Context) error {
	fmt.Printf("%+v\n", context.Pack())
	fmt.Printf("connection[%d]", context.Connection().FD())
	return context.Connection().Send(context.Pack())
}

func (h *Handler) Close(conn connection.IConnection) error {
	fmt.Printf("connection[%d] close \n", conn.FD())
	return nil
}

func (h *Handler) Packet(buf []byte) (connection.IPacket, error) {
	p := &Packet{}
	err := p.Unserialize(buf)
	if err != nil {
		return nil, err
	}

	return p, nil
}

type CHandler struct {
}

func (h *CHandler) Packet(buf []byte) (connection.IPacket, error) {
	p := &Packet{}
	err := p.Unserialize(buf)
	if err != nil {
		fmt.Println("buf: ", buf, "err:", err)
		return nil, err
	}

	return p, nil
}

func (h *CHandler) Receive(pack connection.IPacket, cli *client.Client) error {
	fmt.Printf("%+v\n", pack)
	time.AfterFunc(10*time.Second, func() {
		if err := cli.Send(pack); err != nil {
			debug.Erro("send failure, error: %s", err)
		}
	})

	return nil
}

func (h *CHandler) Idle(cli *client.Client) error {
	return nil
}

func (h *CHandler) Try(cli *client.Client) bool {
	return false
}

func (h *CHandler) Shutdown() {
	fmt.Println("client Shutdown")
}
