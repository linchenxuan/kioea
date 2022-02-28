package tcp_service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"sandbox/blogo/pkg/proto/pmonitor"
	"testing"
	"time"
)

func TestServer_Start(t *testing.T) {
	server := NewServer(ServerOption{
		TcpOption: TcpOption{
			IP:               "127.0.0.1",
			Port:             9765,
			SendBufferSize:   1024,
			HandshakeTimeout: time.Second,
		},
		MsgHandlerOption: MsgHandlerOption{
			WorkerPoolSize:     20,
			MaxTaskQueueLen:    30,
			FindWorkerRetry:    5,
			FindWorkerWaitTime: time.Millisecond * 10,
		},
	})
	server.AddRouter(pmonitor.GMT_GameHeartbeat, &testRouter{})
	server.Start()
	time.Sleep(time.Second)

	go client_send(0, 10000)
	go client_send(10000, 20000)
	select {}
}

type testRouter struct {
	BaseRouter
}

func (r *testRouter) Handler(request IRequest) {
	//time.Sleep(time.Minute)
	//request.GetConnection().SendMsg(1, []byte("2111111111"))
	log.Info("recv: ", string(request.GetData()))
}

func client_send(start, end int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", 9765))
	if err != nil {
		log.Error(err)
		return
	}

	dp := NewDataPacket()
	for i := start; i < end; i++ {
		m := pmonitor.GameHeartbeat{UserNumber: i}
		d, err := m.Marshal()
		if err != nil {
			log.Error(err)
			return
		}
		err = dp.Packet(conn, &Messages{
			Type: 4,
			Data: d,
		})
		if err != nil {
			log.Error(err)
			return
		}
		time.Sleep(time.Millisecond)
	}
}
