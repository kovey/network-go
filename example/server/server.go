package main

import (
	"network/connection"
	"network/example"
	"network/server"
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
	serv.SetService(server.NewTcpService(1024))
	serv.SetHandler(&example.Handler{})
	serv.Run()
}
