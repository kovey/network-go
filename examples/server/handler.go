package main

import (
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/network-go/v2/connection"
	"github.com/kovey/network-go/v2/server"
)

type data struct {
	Id      int32   `json:"id"`
	Name    string  `json:"name"`
	Ok      bool    `json:"ok"`
	Balance float32 `json:"balance"`
}

type handler struct {
}

func (h *handler) Connect(conn *connection.Connection) error {
	debug.Info("new connection: %d", conn.FD())
	return nil
}

func (h *handler) Receive(ctx *server.Context) error {
	var dt []data
	if err := json.Unmarshal(ctx.Data.Body, &dt); err != nil {
		return err
	}

	debug.Info("data: %+v", dt)
	ctx.Conn.Write(append(ctx.Data.Header, ctx.Data.Body...))
	return nil
}

func (h *handler) Close(conn *connection.Connection) error {
	debug.Info("close connection: %d", conn.FD())
	return nil
}

func main() {
	tcp := server.NewTcpService(1024)
	tcp.WithBodyLenOffset(0).WithBodyLenghLen(4).WithEndian(binary.BigEndian).WithHeaderLenType(connection.Len_Type_Int32).WithMaxLen(81290).WithMaxIdleTime(10 * time.Second)
	serv := server.NewServer("0.0.0.0", 9910)
	serv.SetHandler(&handler{})
	serv.SetService(tcp)
	serv.Run()
}
