package tcp_service

type MsgType int32

func (m MsgType) Len() int {
	return 4
}

type IMessages interface {
	GetMsgType() MsgType
	GetMsgData() []byte
	Len() int32
}

type Messages struct {
	Type MsgType
	Data []byte
}

func (m *Messages) GetMsgType() MsgType {
	return m.Type
}

func (m *Messages) GetMsgData() []byte {
	return m.Data
}
func (m *Messages) Len() int32 {
	return int32(m.GetMsgType().Len() + len(m.GetMsgData()))
}
