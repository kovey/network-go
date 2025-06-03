package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/debug-go/run"
	"github.com/kovey/network-go/v2/connection"
)

type IService interface {
	Listen(host string, port int) error
	Accept() (*connection.Connection, error)
	Close()
	Shutdown()
	IsClosed() bool
}

type IHandler interface {
	Connect(*connection.Connection) error
	Receive(*Context) error
	Close(*connection.Connection) error
}

type Server struct {
	conns      sync.Map
	service    IService
	handler    IHandler
	wait       sync.WaitGroup
	isMaintain bool
	host       string
	port       int
	OnSuccess  func(*Server)
}

func NewServer(host string, port int) *Server {
	return &Server{conns: sync.Map{}, wait: sync.WaitGroup{}, host: host, port: port, isMaintain: false}
}

func (s *Server) WithService(service IService) *Server {
	s.service = service
	return s
}

func (s *Server) WithHandler(handler IHandler) *Server {
	s.handler = handler
	return s
}

func (s *Server) listenAndServ() error {
	return s.service.Listen(s.host, s.port)
}

func (s *Server) connect(conn *connection.Connection) {
	defer func() {
		run.Panic(recover())
	}()
	if err := s.handler.Connect(conn); err != nil {
		debug.Erro("connection[%d] on connect failure, error: %s", conn.FD(), err)
	}
}

func (s *Server) loop() {
	for {
		if s.service.IsClosed() {
			debug.Erro("service closed")
			break
		}

		conn, err := s.service.Accept()
		if err != nil {
			debug.Erro("accept error: %s", err)
			continue
		}

		if s.isMaintain {
			conn.Close()
			debug.Erro("server is into maintain")
			continue
		}

		s.conns.Store(conn.FD(), conn)
		s.wait.Add(1)
		go s.handlerConn(conn)
	}
	debug.Warn("server main loop exit")
}

func (s *Server) Close(fd uint64) error {
	conn, ok := s.conns.Load(fd)
	if !ok {
		return nil
	}
	s.conns.Delete(fd)
	c, sure := conn.(*connection.Connection)
	if !sure {
		return nil
	}

	s.service.Close()

	err := c.Close()
	if err != nil {
		return err
	}

	return s.handler.Close(c)
}

func (s *Server) Send(pack []byte, fd int) error {
	if pack == nil {
		return fmt.Errorf("pack is empty")
	}

	conn, ok := s.conns.Load(fd)
	if !ok {
		return fmt.Errorf("connection[%d] is not exists", fd)
	}

	c, sure := conn.(*connection.Connection)
	if !sure {
		return fmt.Errorf("connection[%d] is not implements connection.IConnection", fd)
	}

	return c.Write(pack)
}

func (s *Server) Shutdown() {
	s.service.Shutdown()
	s.conns.Range(func(fd, conn interface{}) bool {
		id, ok := fd.(uint64)
		if !ok {
			return true
		}

		s.Close(id)
		return true
	})

	s.wait.Wait()
	s.handler = nil
}

func (s *Server) handlerConn(conn *connection.Connection) {
	s.connect(conn)
	defer s.wait.Done()
	defer func() {
		run.Panic(recover())
	}()
	defer s.Close(conn.FD())
	go conn.ReadLoop()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		if s.isMaintain {
			debug.Erro("server is into maintain")
			continue
		}

		select {
		case pbuf, ok := <-conn.Packets():
			if !ok {
				return
			}

			s.handlerPacket(pbuf, conn)
		case now := <-ticker.C:
			if conn.Expired(now) {
				return
			}
		}

	}
}

func (s *Server) handlerPacket(data *connection.Packet, conn *connection.Connection) {
	defer func() {
		run.Panic(recover())
	}()
	context := NewContext(context.Background())
	defer context.Drop()

	context.Conn = conn
	context.Data = data

	if err := s.handler.Receive(context); err != nil {
		debug.Erro("handler receive error: %s", err)
	}
}

func (s *Server) ListenAndServ() {
	err := s.listenAndServ()
	if err != nil {
		panic(err)
	}

	if s.OnSuccess != nil {
		s.OnSuccess(s)
	}

	s.loop()
}

func (s *Server) Maintain() {
	s.isMaintain = true
}
