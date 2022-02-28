package tcp_service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"runtime"
	"time"
)

type IMsgHandler interface {
	AddRouter(MsgType, IRouter)
	DoMsgHandler(IRequest)
	StartWorkerPool()
	SendMsgToTaskQueue(IRequest)
}

type MsgHandler struct {
	Handlers           map[MsgType]IRouter
	WorkerPoolSize     int
	MaxTaskQueueLen    int
	TaskQueue          []chan IRequest
	FindWorkerWaitTime time.Duration
	FindWorkerRetry    int
	r                  *rand.Rand
}

func NewMsgHandler(option MsgHandlerOption) *MsgHandler {
	return &MsgHandler{
		WorkerPoolSize:     option.WorkerPoolSize,
		MaxTaskQueueLen:    option.MaxTaskQueueLen,
		Handlers:           make(map[MsgType]IRouter),
		TaskQueue:          make([]chan IRequest, option.WorkerPoolSize),
		FindWorkerWaitTime: option.FindWorkerWaitTime,
		FindWorkerRetry:    option.FindWorkerRetry,
		r:                  rand.New(rand.NewSource(time.Now().Unix())),
	}

}
func (m *MsgHandler) AddRouter(t MsgType, router IRouter) {
	if _, ok := m.Handlers[t]; ok {
		return
	}
	m.Handlers[t] = router
}

func (m *MsgHandler) DoMsgHandler(request IRequest) {
	handler, ok := m.Handlers[request.GetMsgType()]
	if !ok {
		log.WithFields(log.Fields{
			"MsgType": request.GetMsgType(),
		}).Error("can not find this msg type handler")
		return
	}

	defer func() {
		if err := recover(); err != nil {
			const size = 16 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.WithFields(log.Fields{
				"stack":   fmt.Sprintf("\n%s", buf),
				"error":   err,
				"msgType": request.GetMsgType(),
				"message": string(request.GetData()),
			}).Panic("MsgHandler panic error")
		}
	}()

	handler.PreHandler(request)
	handler.Handler(request)
	handler.PostHandler(request)
}

func (m *MsgHandler) StartWorkerPool() {
	for i := 0; i < m.WorkerPoolSize; i++ {
		m.TaskQueue[i] = make(chan IRequest, m.MaxTaskQueueLen)
		go m.StartOneWorker(m.TaskQueue[i])
	}
}

func (m *MsgHandler) StartOneWorker(workerQueue chan IRequest) {
	for true {
		select {
		case req := <-workerQueue:
			m.DoMsgHandler(req)
		}
	}
}

func (m *MsgHandler) SendMsgToTaskQueue(req IRequest) {
	ch := m.getTaskQueue()
	ch <- req
}

func (m *MsgHandler) getTaskQueue() chan IRequest {
	index, _l := 0, m.MaxTaskQueueLen
	for i := 0; i <= m.FindWorkerRetry; i++ {
		for j, que := range m.TaskQueue {
			l := len(que)
			if l == 0 {
				return que
			} else if l < _l {
				_l = l
				index = j
			}
		}
		if _l != m.MaxTaskQueueLen {
			break
		}
		time.Sleep(m.FindWorkerWaitTime)
	}
	if _l == m.MaxTaskQueueLen {
		index = int(m.r.Int31n(int32(m.WorkerPoolSize)))
		log.WithFields(log.Fields{
			"MaxTaskQueueLen":    m.MaxTaskQueueLen,
			"WorkerPoolSize":     m.WorkerPoolSize,
			"FindWorkerRetry":    m.FindWorkerRetry,
			"FindWorkerWaitTime": m.FindWorkerWaitTime,
		}).Debug("all task queues are full , find a random queue")
	}
	return m.TaskQueue[index]
}
