package tcp_service

import (
	"encoding/binary"
	"fmt"
	"io"
)

const DefaultMaxMsgLen = 128 * 1024

type IDataPacket interface {
	Packet(io.Writer, IMessages) error
	UnPacket(io.Reader) (IMessages, error)
}

type DataPacket struct {
	MaxMsgLen int32
}

func NewDataPacket() *DataPacket {
	return &DataPacket{DefaultMaxMsgLen}
}

func (packet *DataPacket) Packet(w io.Writer, msg IMessages) error {
	_l := msg.Len()
	if err := binary.Write(w, binary.BigEndian, _l); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, msg.GetMsgType()); err != nil {
		return err
	}

	return binary.Write(w, binary.BigEndian, msg.GetMsgData())
}

func (packet *DataPacket) UnPacket(r io.Reader) (IMessages, error) {
	var msg Messages
	// 读长度
	length := msg.Len()
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	if int(length) < msg.Type.Len() {
		return nil, fmt.Errorf("UnPacket error lenth")
	}

	if length > packet.MaxMsgLen {
		return nil, fmt.Errorf("msg too long,the data packet support max msg lenth is %d,recved msg lenth is %d", packet.MaxMsgLen, length)
	}
	// 读消息类型
	err = binary.Read(r, binary.BigEndian, &msg.Type)
	if err != nil {
		return nil, err
	}
	dataLen := int(length) - msg.Type.Len()
	// 读消息内容
	data := make([]byte, dataLen)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}
	msg.Data = data
	return &msg, nil
}
