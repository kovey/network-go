package server

import (
	"fmt"
	"network/connection"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

type Context struct {
	conn connection.IConnection
	pack connection.IPacket
}

func (c *Context) SetConnection(conn connection.IConnection) {
	c.conn = conn
}

func (c *Context) SetPack(pack connection.IPacket) {
	c.pack = pack
}

func (c *Context) Connection() connection.IConnection {
	return c.conn
}

func (c *Context) Pack() connection.IPacket {
	return c.pack
}

func (c *Context) Reset() {
	c.conn = nil
	c.pack = nil
}

func putContext(c *Context) {
	c.Reset()
	pool.Put(c)
}

func getContext() (*Context, error) {
	context, ok := pool.Get().(*Context)
	if !ok {
		return nil, fmt.Errorf("context is not exists")
	}

	return context, nil
}
