package client

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/debug-go/run"
	"github.com/kovey/network-go/connection"
)

type IClient interface {
	Dial(host string, port int) error
	Connection() connection.IConnection
}

type IHandler interface {
	Packet([]byte) (connection.IPacket, error)
	Receive(connection.IPacket, *Client) error
	Idle(*Client) error
	Try(*Client) bool
	Shutdown()
}

type Client struct {
	cli        IClient
	handler    IHandler
	wait       sync.WaitGroup
	config     connection.PacketConfig
	shutdown   chan bool
	ticker     *time.Ticker
	host       string
	port       int
	isShutdown bool
}

func NewClient(config connection.PacketConfig) *Client {
	return &Client{wait: sync.WaitGroup{}, config: config, shutdown: make(chan bool, 1), ticker: time.NewTicker(10 * time.Second), isShutdown: false}
}

func (c *Client) SetService(cli IClient) {
	c.cli = cli
}

func (c *Client) SetHandler(handler IHandler) {
	c.handler = handler
}

func (c *Client) Dial(host string, port int) error {
	c.host = host
	c.port = port
	return c.cli.Dial(host, port)
}

func (c *Client) handlerPacket(pack connection.IPacket) {
	defer c.wait.Done()
	defer func() {
		run.Panic(recover())
	}()
	if err := c.handler.Receive(pack, c); err != nil {
		debug.Erro("connection[%d] on receive failure, error: %s", c.cli.Connection().FD(), err)
	}
}

func (c *Client) Try() error {
	return c.Dial(c.host, c.port)
}

func (c *Client) rloop() {
	defer c.wait.Done()
	defer func() {
		run.Panic(recover())
	}()
	for {
		if c.isShutdown {
			break
		}

		pbuf, err := c.cli.Connection().Read(c.config.HeaderLength, c.config.BodyLenLen, c.config.BodyLenOffset)
		if c.isShutdown {
			break
		}

		if err == io.EOF {
			if !c.handler.Try(c) {
				c.shutdown <- true
				break
			}

			continue
		}

		if err != nil {
			if !c.handler.Try(c) {
				c.shutdown <- true
				debug.Erro("connection[%d] read data failure, error: %s", c.cli.Connection().FD(), err)
				break
			}

			continue
		}

		if c == nil || c.handler == nil {
			debug.Erro("client is nil or handler is nil")
			break
		}

		pack, e := c.handler.Packet(pbuf)
		if e != nil {
			debug.Erro("packet failure, error: %s", e)
			continue
		}
		if pack == nil {
			debug.Erro("pack is nil")
			continue
		}

		c.cli.Connection().RQueue() <- pack
	}
}

func (c *Client) Loop() {
	c.wait.Add(1)
	go c.rloop()

	for {
		select {
		case <-c.shutdown:
			c.Close()
			c.isShutdown = true
			return
		case pack, ok := <-c.cli.Connection().RQueue():
			if !ok {
				c.Close()
				c.isShutdown = true
				return
			}
			c.wait.Add(1)
			go c.handlerPacket(pack)
		case pack, ok := <-c.cli.Connection().WQueue():
			if !ok {
				c.Close()
				c.isShutdown = true
				return
			}
			_, err := c.cli.Connection().Write(pack)
			if err != nil {
				debug.Erro("write pack to connection failure, error: %s", err)
			}
		case <-c.ticker.C:
			c.wait.Add(1)
			go c.handlerIdle()
		}
	}
}

func (c *Client) handlerIdle() {
	defer c.wait.Done()
	defer func() {
		run.Panic(recover())
	}()
	if err := c.handler.Idle(c); err != nil {
		debug.Erro("Idle failure, error: %s", err)
	}
}

func (c *Client) Close() {
	c.handler.Shutdown()
	c.cli.Connection().Close()
	c.ticker.Stop()
	c.wait.Wait()
}

func (c *Client) Shutdown() {
	if c.isShutdown {
		return
	}
	c.shutdown <- true
	c.isShutdown = true
}

func (c *Client) Send(pack connection.IPacket) error {
	if pack == nil {
		return fmt.Errorf("pack is empty")
	}

	buf := pack.Serialize()
	if buf == nil {
		return fmt.Errorf("pack is empty")
	}

	return c.SendBytes(buf)
}

func (c *Client) SendBytes(buf []byte) error {
	if c.cli.Connection().Closed() {
		return fmt.Errorf("connection[%d] is closed", c.cli.Connection().FD())
	}

	c.cli.Connection().WQueue() <- buf
	return nil
}
