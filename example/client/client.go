package main

import (
	"github.com/kovey/network-go/client"
	"github.com/kovey/network-go/connection"
	"github.com/kovey/network-go/example"
)

func main() {
	cli := client.NewClient(connection.PacketConfig{HeaderLength: 4, BodyLenOffset: 0, BodyLenLen: 4})
	cli.SetService(client.NewTcp())
	cli.SetHandler(&example.CHandler{})
	err := cli.Dial("127.0.0.1", 9911)
	if err != nil {
		panic(err)
	}

	pack := &example.Packet{}
	pack.Action = 1000
	pack.Name = "kovey"
	pack.Age = 18

	if err := cli.Send(pack); err != nil {
		panic(err)
	}

	cli.Loop()
}
