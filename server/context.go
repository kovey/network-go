package server

import (
	"fmt"
	"github.com/kovey/network-go/connection"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

type Context struct {
	conn    connection.IConnection
	pack    connection.IPacket
	traceId string
	spanId  string
}

func (c *Context) SetTraceId(traceId string) {
	c.traceId = traceId
}

func (c *Context) SetSpanId(spanId string) {
	c.spanId = spanId
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

func (c *Context) TraceId() string {
	return c.traceId
}

func (c *Context) SpanId() string {
	return c.spanId
}

func (c *Context) Reset() {
	c.conn = nil
	c.pack = nil
	c.traceId = ""
	c.spanId = ""
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
