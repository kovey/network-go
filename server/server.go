package server

import (
	"fmt"
	"io"
	"sync"

	"github.com/kovey/network-go/connection"

	"github.com/kovey/logger-go/logger"
)

type IService interface {
	Listen(host string, port int) error
	Accept() (connection.IConnection, error)
	Close()
	Shutdown()
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
		logger.Panic(recover())
	}()
	s.handler.Connect(conn)
}

func (s *Server) loop() {
	for {
		conn, err := s.service.Accept()
		if err != nil {
			break
		}

		if s.isMaintain {
			conn.Close()
			logger.Debug("server is into maintain")
			continue
		}

		s.conns.Store(conn.FD(), conn)
		s.wait.Add(1)
		go s.connect(conn)
		s.wait.Add(1)
		go s.handlerConn(conn)
	}
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

	select {
	case c.WQueue() <- buf:
		return nil
	}
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
		logger.Panic(recover())
	}()
	defer s.Close(conn.FD())
	for {
		pbuf, err := conn.Read(s.config.PConfig.HeaderLength, s.config.PConfig.BodyLenLen, s.config.PConfig.BodyLenOffset)
		if err == io.EOF {
			break
		}

		if err != nil {
			break
		}

		if s.isMaintain {
			logger.Debug("server is into maintain")
			continue
		}

		pack, e := s.handler.Packet(pbuf)
		if e != nil || pack == nil {
			logger.Error("get packet error or packet is nil: %s, %+v", e, pack)
			continue
		}

		if conn.Closed() {
			break
		}

		select {
		case conn.RQueue() <- pack:
		}
	}
}

func (s *Server) handlerConn(conn connection.IConnection) {
	defer s.wait.Done()
	defer func() {
		logger.Panic(recover())
	}()
	s.wait.Add(1)
	go s.rloop(conn)

conn_loop:
	for {
		select {
		case pack, ok := <-conn.RQueue():
			if !ok {
				break conn_loop
			}
			if pack == nil {
				continue conn_loop
			}
			s.wait.Add(1)
			go s.handlerPacket(pack, conn)
		case pack, ok := <-conn.WQueue():
			if !ok {
				break conn_loop
			}
			if pack == nil {
				continue conn_loop
			}

			n, err := conn.Write(pack)
			logger.Debug("send data result, n[%d], err[%s]", n, err)
		}
	}
}

func (s *Server) handlerPacket(pack connection.IPacket, conn connection.IConnection) {
	defer s.wait.Done()
	defer func() {
		logger.Panic(recover())
	}()
	context, err := getContext()
	if err != nil {
		logger.Error("get context error: %s", err)
		return
	}
	defer putContext(context)

	context.SetConnection(conn)
	context.SetPack(pack)

	err = s.handler.Receive(context)
	if err != nil {
		logger.Error("handler receive error: %s", err)
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
