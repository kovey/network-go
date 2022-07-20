package client

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kovey/network-go/connection"

	"github.com/kovey/logger-go/logger"
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
		logger.Panic(recover())
	}()
	c.handler.Receive(pack, c)
}

func (c *Client) Try() error {
	return c.Dial(c.host, c.port)
}

func (c *Client) rloop() {
	defer c.wait.Done()
	defer func() {
		logger.Panic(recover())
	}()
	for {
		pbuf, err := c.cli.Connection().Read(c.config.HeaderLength, c.config.BodyLenLen, c.config.BodyLenOffset)
		if err == io.EOF {
			if !c.handler.Try(c) {
				c.shutdown <- true
				break
			}

			continue
		}

		if err != nil {
			c.shutdown <- true
			break
		}

		if c == nil || c.handler == nil {
			break
		}

		pack, e := c.handler.Packet(pbuf)
		if e != nil || pack == nil {
			continue
		}

		select {
		case c.cli.Connection().RQueue() <- pack:
		}
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
			c.cli.Connection().Write(pack)
		case <-c.ticker.C:
			c.wait.Add(1)
			go c.handlerIdle()
		}
	}
}

func (c *Client) handlerIdle() {
	defer c.wait.Done()
	defer func() {
		logger.Panic(recover())
	}()
	c.handler.Idle(c)
}

func (c *Client) Close() {
	c.handler.Shutdown()
	c.handler = nil
	c.cli.Connection().Close()
	c.ticker.Stop()
	close(c.shutdown)
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
	if c.cli.Connection().Closed() {
		return fmt.Errorf("connection[%d] is closed", c.cli.Connection().FD())
	}
	if pack == nil {
		return fmt.Errorf("pack is empty")
	}

	buf := pack.Serialize()
	if buf == nil {
		return fmt.Errorf("pack is empty")
	}

	select {
	case c.cli.Connection().WQueue() <- buf:
		return nil
	}
}
