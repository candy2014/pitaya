package message

import (
	"encoding/binary"
	"github.com/topfreegames/pitaya/conn/codec"
	"strconv"
)

type ForwardMessageEncoder struct {
	DataCompression bool
}

func NewForwardMessageEncoder(dataCompression bool) *ForwardMessageEncoder {
	return &ForwardMessageEncoder{
		dataCompression,
	}
}

// IsCompressionEnabled returns wether the compression is enabled or not
func (f *ForwardMessageEncoder) IsCompressionEnabled() bool {
	return f.DataCompression
}

func (f *ForwardMessageEncoder) Encode(message *Message) ([]byte, error) {
	data := message.Data
	length := 6 + len(data)
	sendData := make([]byte, length)
	// int8
	sendData[0] = 4
	copy(sendData[1:codec.HeadLength], codec.IntToBytes(length))
	sendData[length-2] = 0xEE
	sendData[length-1] = 0xEE
	copy(sendData[codec.HeadLength:], data)
	return sendData, nil
}

// ForwardDecode decodes the message
// 把网关的packet内的[]byte转成rpc内的消息
func ForwardDecode(data []byte) (*Message, error) {
	if len(data) < 4 {
		return nil, ErrInvalidMessage
	}
	offset := 0
	//读取cmd 然后配置路由信息
	cmdId := binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	m := New()
	m.ID = 1
	route := strconv.Itoa(int(cmdId))
	m.Route = route
	m.Data = data[offset:]
	m.Err = false
	m.compressed = false
	return m, nil
}
