package tcp_service

import "time"

type ServerOption struct {
	TcpOption
	MsgHandlerOption
}
type TcpOption struct {
	IP               string
	Port             int
	SendBufferSize   int
	HandshakeTimeout time.Duration
}

type MsgHandlerOption struct {
	WorkerPoolSize     int
	MaxTaskQueueLen    int
	FindWorkerWaitTime time.Duration
	FindWorkerRetry    int
}
