package server

import (
	"context"

	"github.com/kovey/network-go/v2/connection"
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
	Conn    *connection.Connection
	Data    *connection.Packet
	TraceId string
	SpanId  string
}

func NewContext(parent context.Context) *Context {
	pc := pool.NewContext(parent)
	ctx := pc.GetNoCtx(ctx_namespace, ctx_name).(*Context)
	ctx.Context = pc
	return ctx
}

func (c *Context) Reset() {
	c.Conn = nil
	c.Data = nil
	c.TraceId = ""
	c.SpanId = ""
	c.Context = nil
}
