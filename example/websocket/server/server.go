package main

import (
	"github.com/kovey/network-go/connection"
	"github.com/kovey/network-go/example/websocket"
	"github.com/kovey/network-go/server"
)

func main() {
	config := server.Config{}
	packet := connection.PacketConfig{}
	packet.HeaderLength = 4
	packet.BodyLenOffset = 0
	packet.BodyLenLen = 4
	config.PConfig = packet
	config.Host = "127.0.0.1"
	config.Port = 9911

	serv := server.NewServer(config)
	serv.SetService(server.NewWebSocketService(1024))
	serv.SetHandler(&websocket.Handler{})
	serv.Run()
}
