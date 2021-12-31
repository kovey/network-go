package client

import (
	"fmt"
	"net"
	"network/connection"
)

type Tcp struct {
	conn *connection.Tcp
}

func NewTcp() *Tcp {
	return &Tcp{}
}

func (t *Tcp) Dial(host string, port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	t.conn = connection.NewTcp(1, conn)
	return nil
}

func (t *Tcp) Connection() connection.IConnection {
	return t.conn
}
