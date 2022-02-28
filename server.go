package tcp_service

import (
	"fmt"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type IServer interface {
	Start()
	Stop()
	AddRouter(MsgType, IRouter)
	GetConnManager() IConnManager
	SetOnConnStartHook(func(connection IConnection))
	SetOnConnStopHook(func(connection IConnection))
	CallOnConnStartHook(connection IConnection)
	CallOnConnStopHook(connection IConnection)
	GetHandshakeTimeOut() time.Duration
}

type Server struct {
	TcpOption
	MsgHandler  IMsgHandler
	ConnManager IConnManager
	OnConnStart func(connection IConnection)
	OnConnStop  func(connection IConnection)
	Package     IDataPacket
}

func NewServer(option ServerOption) IServer {
	s := &Server{
		TcpOption:   option.TcpOption,
		MsgHandler:  NewMsgHandler(option.MsgHandlerOption),
		ConnManager: NewConnManager(),
		Package:     NewDataPacket(),
	}
	return s
}
func (s *Server) Start() {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		log.WithFields(log.Fields{
			"IP":    s.IP,
			"Port":  s.Port,
			"Error": err,
		}).Error("Server ResolveTCPAddr failed")
		return
	}
	go func() {
		s.MsgHandler.StartWorkerPool()
		listen, err := net.ListenTCP("tcp", addr)
		if err != nil {
			log.WithFields(log.Fields{
				"IP":    s.IP,
				"Port":  s.Port,
				"Error": err,
			}).Error("Server start error")
			return
		}
		for {
			conn, err := listen.AcceptTCP()
			if err != nil {
				log.WithField("error", err).Error("Server accept TCP error")
				continue
			}
			cId := xid.New().String()
			c := NewConnection(s, conn, cId, s.Package, s.MsgHandler, s.SendBufferSize)
			go c.Start()
		}
	}()
}

func (s *Server) Stop() {
	s.ConnManager.ClearConn()
}

func (s *Server) AddRouter(t MsgType, router IRouter) {
	s.MsgHandler.AddRouter(t, router)
}

func (s *Server) GetConnManager() IConnManager {
	return s.ConnManager
}

func (s *Server) SetOnConnStartHook(hookHandler func(connection IConnection)) {
	s.OnConnStart = hookHandler
}

func (s *Server) SetOnConnStopHook(hookHandler func(conn IConnection)) {
	s.OnConnStop = hookHandler
}

func (s *Server) CallOnConnStartHook(conn IConnection) {
	if s.OnConnStart != nil {
		s.OnConnStart(conn)
	}
}

func (s *Server) CallOnConnStopHook(conn IConnection) {
	if s.OnConnStop != nil {
		s.OnConnStop(conn)
	}
}

func (s *Server) GetHandshakeTimeOut() time.Duration {
	return s.HandshakeTimeout
}
