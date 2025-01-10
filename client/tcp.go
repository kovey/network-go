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
	return &Tcp{conn: connection.NewConnection(1, nil)}
}

func (t *Tcp) Dial(host string, port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	t.conn.WithConn(conn)
	return nil
}

func (t *Tcp) Connection() *connection.Connection {
	return t.conn
}

func (t *Tcp) WithHeaderLenType(l connection.HeaderLenType) *Tcp {
	t.conn.WithHeaderLenType(l)
	return t
}

func (t *Tcp) WithEndian(e binary.ByteOrder) *Tcp {
	t.conn.WithEndian(e)
	return t
}

func (t *Tcp) WithMaxLen(maxLen int) *Tcp {
	t.conn.WithMaxLen(maxLen)
	return t
}

func (t *Tcp) WithBodyLenghLen(length int) *Tcp {
	t.conn.WithBodyLenghLen(length)
	return t
}

func (t *Tcp) WithBodyLenOffset(offset int) *Tcp {
	t.conn.WithBodyLenOffset(offset)
	return t
}
