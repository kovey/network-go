package server

import (
	"context"

	"github.com/kovey/network-go/connection"
	"github.com/kovey/pool"
	"github.com/kovey/pool/object"
)

const (
	ctx_namespace = "ko.network.context"
	ctx_name      = "Context"
)

func init() {
	pool.DefaultNoCtx(ctx_namespace, ctx_name, func() any {
		return &Context{ObjNoCtx: object.NewObjNoCtx(ctx_namespace, ctx_name)}
	})
}

type Context struct {
	*object.ObjNoCtx
	*pool.Context
	conn    connection.IConnection
	pack    connection.IPacket
	traceId string
	spanId  string
}

func NewContext(parent context.Context) *Context {
	pc := pool.NewContext(parent)
	ctx := pc.GetNoCtx(ctx_namespace, ctx_name).(*Context)
	ctx.Context = pc
	return ctx
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
	c.Context = nil
}
