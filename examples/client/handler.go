package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/network-go/v2/client"
	"github.com/kovey/network-go/v2/connection"
)

type data struct {
	Id      int32   `json:"id"`
	Name    string  `json:"name"`
	Ok      bool    `json:"ok"`
	Balance float32 `json:"balance"`
}

type handler struct {
}

func (h *handler) Receive(packet *connection.Packet, cli *client.Client) error {
	var dt []data
	if err := json.Unmarshal(packet.Body, &dt); err != nil {
		return err
	}

	debug.Info("data: %+v", dt)
	cli.Send(append(packet.Header, packet.Body...))
	return nil
}

func (h *handler) Idle(cli *client.Client) error {
	return nil
}

func (h *handler) Try(cli *client.Client) bool {
	if err := cli.Redial(); err == nil {
		return true
	}

	ticker := time.NewTicker(5 * time.Second)
	count := 0
	for {
		<-ticker.C
		if err := cli.Redial(); err == nil {
			return true
		}

		count++
		if count >= 10 {
			break
		}
	}

	return false
}

func (h *handler) Shutdown() {
}

func main() {
	tcp := client.NewTcp()
	tcp.WithBodyLenOffset(0).WithBodyLengthLen(4).WithEndian(binary.BigEndian).WithHeaderLenType(connection.Len_Type_Int32).WithMaxLen(81920)
	cli := client.NewClient().WithHandler(&handler{}).WithService(tcp)
	if err := cli.Dial("127.0.0.1", 9910); err != nil {
		panic(err)
	}

	var dt []data
	for i := 0; i < 1000; i++ {
		dt = append(dt, data{Id: 1000 + int32(i), Name: fmt.Sprintf("kovey_%d", i), Ok: i%2 == 1, Balance: 1000000})
	}
	buff, _ := json.Marshal(dt)
	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, int32(len(buff)))
	cli.Send(append(b.Bytes(), buff...))
	cli.Listen()
}
