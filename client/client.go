package client

import (
	"io"
	"sync"
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/debug-go/run"
	"github.com/kovey/network-go/v2/connection"
)

type IClient interface {
	Dial(host string, port int) error
	Connection() *connection.Connection
}

type IHandler interface {
	Receive([]byte, *Client) error
	Idle(*Client) error
	Try(*Client) bool
	Shutdown()
}

type Client struct {
	cli        IClient
	handler    IHandler
	wait       sync.WaitGroup
	shutdown   chan bool
	ticker     *time.Ticker
	host       string
	port       int
	isShutdown bool
}

func NewClient() *Client {
	return &Client{wait: sync.WaitGroup{}, shutdown: make(chan bool, 1), ticker: time.NewTicker(10 * time.Second), isShutdown: false}
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

func (c *Client) handlerPacket(data []byte) {
	defer func() {
		run.Panic(recover())
	}()
	if err := c.handler.Receive(data, c); err != nil {
		debug.Erro("connection[%d] on receive failure, error: %s", c.cli.Connection().FD(), err)
	}
}

func (c *Client) Try() error {
	return c.Dial(c.host, c.port)
}

func (c *Client) Loop() {
	defer func() {
		run.Panic(recover())
	}()
	c.wait.Add(1)
	go c.handlerIdle()

	for {
		if c.isShutdown {
			break
		}

		pbuf, err := c.cli.Connection().Read()
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

		c.handlerPacket(pbuf)
	}
}

func (c *Client) handlerIdle() {
	defer c.wait.Done()
	defer func() {
		run.Panic(recover())
	}()

	for {
		select {
		case <-c.shutdown:
			return
		case <-c.ticker.C:
			if err := c.handler.Idle(c); err != nil {
				debug.Erro("Idle failure, error: %s", err)
			}
		}
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
	c.cli.Connection().Close()
}

func (c *Client) Send(data []byte) error {
	return c.cli.Connection().Write(data)
}
