package main

import (
	"github.com/kovey/network-go/client"
	"github.com/kovey/network-go/connection"
	"github.com/kovey/network-go/example/websocket"

	"github.com/kovey/logger-go/logger"
)

func main() {
	logger.SetLevel(logger.LOGGER_INFO)
	cli := client.NewClient(connection.PacketConfig{HeaderLength: 4, BodyLenOffset: 0, BodyLenLen: 4})
	cli.SetService(client.NewWebSocket())
	cli.SetHandler(&websocket.CHandler{})
	err := cli.Dial("127.0.0.1", 9911)
	logger.Debug("dial error: %s", err)
	if err != nil {
		panic(err)
	}

	pack := &websocket.Packet{}
	pack.Action = 1000
	pack.Name = "kovey"
	pack.Age = 18

	err = cli.Send(pack)
	logger.Debug("send error: %s", err)

	cli.Loop()
}