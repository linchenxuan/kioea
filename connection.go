package tcp_service

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type IConnection interface {
	Start()
	Stop()
	GetConnID() string
	GetRemoteInfo() net.Addr
	TrySendMsg(MsgType, []byte) error
	SendMsg(MsgType, []byte)
	SetProperty(key string, value interface{})
	GetProperty(key string) interface{}
	RemoveProperty(key string)
	Handshake()
}

type HandlerFunc func(*net.TCPConn, []byte, int) error

type Connection struct {
	TcpServer  IServer
	ConnID     string
	ctx        context.Context
	cancel     context.CancelFunc
	Conn       *net.TCPConn
	MsgHandler IMsgHandler
	MsgChan    chan Messages
	Packet     IDataPacket
	Property   map[string]interface{}
	locker     sync.RWMutex
	handshake  atomic.Value
}

func NewConnection(server IServer, conn *net.TCPConn, connId string, packet IDataPacket, handler IMsgHandler, sendBufferSize int) IConnection {
	cx, cf := context.WithCancel(context.TODO())
	c := &Connection{
		ctx:        cx,
		cancel:     cf,
		TcpServer:  server,
		ConnID:     connId,
		MsgHandler: handler,
		Conn:       conn,
		MsgChan:    make(chan Messages, sendBufferSize),
		Packet:     packet,
		Property:   make(map[string]interface{}),
		locker:     sync.RWMutex{},
		handshake:  atomic.Value{},
	}
	c.handshake.Store(false)
	c.TcpServer.GetConnManager().AddConn(c)
	return c
}

func (c *Connection) StartReader() {
	defer c.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := c.Packet.UnPacket(c.Conn)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.WithField("error", err).Error("connection reader unpacket error!")
				continue
			}
			req := Request{
				Conn: c,
				Msg:  msg,
			}
			c.MsgHandler.SendMsgToTaskQueue(&req)
		}
	}
}

func (c *Connection) StartWriter() {
	for {
		select {
		case data := <-c.MsgChan:
			err := c.Packet.Packet(c.Conn, &data)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"data":  data,
				}).Error("StartWriter Packet error")
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Connection) Start() {
	go c.StartReader()
	go c.StartWriter()
	go c.monitorHandshake()
	c.TcpServer.CallOnConnStartHook(c)

	select {
	case <-c.ctx.Done():
		c.finalizer()
	}
}

func (c *Connection) monitorHandshake() {
	t := c.TcpServer.GetHandshakeTimeOut()
	if t <= 0 {
		return
	}
	select {
	case <-time.After(t):
		if !c.handshaked() {
			c.Stop()
		}
	}
}

func (c *Connection) finalizer() {
	c.TcpServer.CallOnConnStopHook(c)
	c.Conn.Close()
	c.TcpServer.GetConnManager().RemoveConn(c)
}

func (c *Connection) Stop() {
	c.cancel()
}

func (c *Connection) GetConnID() string {
	return c.ConnID
}

func (c *Connection) GetRemoteInfo() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) SendMsg(t MsgType, data []byte) {
	c.MsgChan <- Messages{Type: t, Data: data}
}

func (c *Connection) TrySendMsg(t MsgType, data []byte) error {
	select {
	case c.MsgChan <- Messages{Type: t, Data: data}:
	default:
		log.WithFields(log.Fields{
			"ConnID":  c.ConnID,
			"MsgType": t,
			"Data":    data,
		}).Warn("SendMsg MsgChan is full,the msg is discarded")
	}
	return errors.New("MsgChan is full")
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.locker.Lock()
	defer c.locker.Unlock()

	c.Property[key] = value
}

func (c *Connection) GetProperty(key string) interface{} {
	c.locker.RLock()
	defer c.locker.RUnlock()

	return c.Property[key]
}

func (c *Connection) RemoveProperty(key string) {
	c.locker.Lock()
	defer c.locker.Unlock()

	delete(c.Property, key)
}

func (c *Connection) Handshake() {
	c.handshake.Store(true)
}

func (c *Connection) handshaked() bool {
	return c.handshake.Load().(bool)
}
