package client

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/kovey/network-go/v2/connection"
)

type Tcp struct {
	conn *connection.Connection
}

func NewTcp() *Tcp {
	return &Tcp{}
}

func (t *Tcp) Dial(host string, port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	t.conn = connection.NewConnection(1, conn)
	return nil
}

func (t *Tcp) Connection() *connection.Connection {
	return t.conn
}

func (t *Tcp) HeaderLenType(l connection.HeaderLenType) *Tcp {
	t.conn.HeaderLenType(l)
	return t
}

func (t *Tcp) Endian(e binary.ByteOrder) *Tcp {
	t.conn.Endian(e)
	return t
}

func (t *Tcp) MaxLen(maxLen int) *Tcp {
	t.conn.MaxLen(maxLen)
	return t
}

func (t *Tcp) BodyLenghLen(length int) *Tcp {
	t.conn.BodyLenghLen(length)
	return t
}

func (t *Tcp) BodyLenOffset(offset int) *Tcp {
	t.conn.BodyLenOffset(offset)
	return t
}
