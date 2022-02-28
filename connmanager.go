package tcp_service

import (
	"errors"
	"sync"
)

type IConnManager interface {
	AddConn(IConnection)
	RemoveConn(connection IConnection)
	GetConnByID(ConnID int) (IConnection, error)
	ClearConn()
	GetConnCount() int
}

type ConnManager struct {
	connMap sync.Map
}

func NewConnManager() *ConnManager {
	return &ConnManager{}
}

func (c *ConnManager) AddConn(conn IConnection) {
	c.connMap.Store(conn.GetConnID(), conn)
}

func (c *ConnManager) RemoveConn(conn IConnection) {
	c.connMap.Delete(conn.GetConnID())
}

func (c *ConnManager) GetConnByID(ConnID int) (IConnection, error) {
	conn, ok := c.connMap.Load(ConnID)
	if !ok {
		return nil, errors.New("connection not found")
	}
	return conn.(IConnection), nil
}

func (c *ConnManager) ClearConn() {
	c.connMap.Range(func(key, value interface{}) bool {
		value.(IConnection).Stop()
		c.connMap.Delete(value.(IConnection).GetConnID())
		return true
	})
}

func (c *ConnManager) GetConnCount() int {
	l := 0
	c.connMap.Range(func(k, v interface{}) bool {
		l++
		return true
	})
	return l
}
