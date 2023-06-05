package server

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/debug-go/run"
	"github.com/kovey/network-go/connection"
)

type IService interface {
	Listen(host string, port int) error
	Accept() (connection.IConnection, error)
	Close()
	Shutdown()
	IsClosed() bool
}

type IHandler interface {
	Connect(connection.IConnection) error
	Receive(*Context) error
	Close(connection.IConnection) error
	Packet([]byte) (connection.IPacket, error)
}

type Server struct {
	conns      sync.Map
	service    IService
	handler    IHandler
	wait       sync.WaitGroup
	config     Config
	isMaintain bool
	OnSuccess  func(*Server)
}

func NewServer(config Config) *Server {
	connection.Init(config.PConfig.Endian)
	return &Server{conns: sync.Map{}, wait: sync.WaitGroup{}, config: config, isMaintain: false}
}

func (s *Server) SetService(service IService) *Server {
	s.service = service
	return s
}

func (s *Server) SetHandler(handler IHandler) *Server {
	s.handler = handler
	return s
}

func (s *Server) listenAndServ() error {
	return s.service.Listen(s.config.Host, s.config.Port)
}

func (s *Server) connect(conn connection.IConnection) {
	defer s.wait.Done()
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
		go s.connect(conn)
		s.wait.Add(1)
		go s.handlerConn(conn)
	}
	debug.Warn("server main loop exit")
}

func (s *Server) Close(fd int) error {
	conn, ok := s.conns.Load(fd)
	if !ok {
		return nil
	}
	s.conns.Delete(fd)
	c, sure := conn.(connection.IConnection)
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

func (s *Server) Send(pack connection.IPacket, fd int) error {
	if pack == nil {
		return fmt.Errorf("pack is empty")
	}

	conn, ok := s.conns.Load(fd)
	if !ok {
		return fmt.Errorf("connection[%d] is not exists", fd)
	}

	c, sure := conn.(connection.IConnection)
	if !sure {
		return fmt.Errorf("connection[%d] is not implements connection.IConnection", fd)
	}

	if c.Closed() {
		return fmt.Errorf("connection[%d] is closed", fd)
	}

	buf := pack.Serialize()
	if buf == nil {
		return fmt.Errorf("pack is nil")
	}

	c.WQueue() <- buf
	return nil
}

func (s *Server) Shutdown() {
	s.service.Shutdown()
	s.conns.Range(func(fd, conn interface{}) bool {
		id, ok := fd.(int)
		if !ok {
			return true
		}

		s.Close(id)
		return true
	})

	s.wait.Wait()
	s.handler = nil
}

func (s *Server) rloop(conn connection.IConnection) {
	defer s.wait.Done()
	defer func() {
		run.Panic(recover())
	}()
	defer s.Close(conn.FD())
	for {
		pbuf, err := conn.Read(s.config.PConfig.HeaderLength, s.config.PConfig.BodyLenLen, s.config.PConfig.BodyLenOffset)
		if err == io.EOF {
			break
		}

		if err != nil {
			debug.Erro("read data error: %s", err)
			break
		}

		if s.isMaintain {
			debug.Erro("server is into maintain")
			continue
		}

		pack, e := s.handler.Packet(pbuf)
		if e != nil || pack == nil {
			debug.Erro("get packet error or packet is nil: %s, %+v", e, pack)
			continue
		}

		if conn.Closed() {
			break
		}

		conn.RQueue() <- pack
	}
}

func (s *Server) handlerConn(conn connection.IConnection) {
	defer s.wait.Done()
	defer func() {
		run.Panic(recover())
	}()
	s.wait.Add(1)
	go s.rloop(conn)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if conn.Expired() {
				debug.Warn("conn[%d] is expired", conn.FD())
				conn.Close()
				return
			}
		case pack, ok := <-conn.RQueue():
			if !ok {
				return
			}
			if pack == nil {
				return
			}
			s.handlerPacket(pack, conn)
		case pack, ok := <-conn.WQueue():
			if !ok {
				return
			}
			if pack == nil {
				return
			}

			n, err := conn.Write(pack)
			debug.Dbug("send data result, n[%d], err[%s]", n, err)
		}
	}
}

func (s *Server) handlerPacket(pack connection.IPacket, conn connection.IConnection) {
	defer func() {
		run.Panic(recover())
	}()
	context, err := getContext()
	if err != nil {
		debug.Erro("get context error: %s", err)
		return
	}
	defer putContext(context)

	context.SetConnection(conn)
	context.SetPack(pack)

	err = s.handler.Receive(context)
	if err != nil {
		debug.Erro("handler receive error: %s", err)
	}
}

func (s *Server) Run() {
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
