package tcp_service

type IRequest interface {
	GetConnection() IConnection
	GetData() []byte
	GetMsgType() MsgType
}

type Request struct {
	Conn IConnection
	Msg  IMessages
}

func (req *Request) GetConnection() IConnection {
	return req.Conn
}

func (req *Request) GetData() []byte {
	return req.Msg.GetMsgData()
}

func (req *Request) GetMsgType() MsgType {
	return req.Msg.GetMsgType()
}
