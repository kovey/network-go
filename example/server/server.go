package main

import (
	"github.com/kovey/logger-go/logger"
	"github.com/kovey/network-go/connection"
	"github.com/kovey/network-go/example"
	"github.com/kovey/network-go/server"
)

func main() {
	logger.SetLevel(logger.LOGGER_INFO)
	config := server.Config{}
	packet := connection.PacketConfig{}
	packet.HeaderLength = 4
	packet.BodyLenOffset = 0
	packet.BodyLenLen = 4
	config.PConfig = packet
	config.Host = "0.0.0.0"
	config.Port = 9911

	serv := server.NewServer(config)
	serv.SetService(server.NewTcpService(1024))
	serv.SetHandler(&example.Handler{})
	serv.Run()
}
