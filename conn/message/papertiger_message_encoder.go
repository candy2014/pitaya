package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

type PapertigerMessageEncoder struct {
	DataCompression bool
}

func NewPapertigerMessageEncoder(dataCompression bool) *PapertigerMessageEncoder {
	return &PapertigerMessageEncoder{
		DataCompression: dataCompression,
	}
}

// IsCompressionEnabled returns wether the compression is enabled or not
func (pme *PapertigerMessageEncoder) IsCompressionEnabled() bool {
	return pme.DataCompression
}

func (pme *PapertigerMessageEncoder) Encode(message *Message) ([]byte, error) {
	bytes.NewBuffer(make([]byte, 1024))
	cmd, _ := strconv.Atoi(message.Route)
	data := message.Data
	length := 7 + len(data)
	sendData := make([]byte, length)
	// int8
	sendData[0] = 101
	binary.LittleEndian.PutUint32(sendData[1:], uint32(cmd))
	sendData[length-2] = 0xEE
	sendData[length-1] = 0xEE
	copy(sendData[5:], data)
	return sendData, nil
}

// DecodePaperTiger decodes the message
func DecodePaperTiger(data []byte) (*Message, error) {
	if len(data) < 4 {
		return nil, ErrInvalidMessage
	}
	offset := 0
	cmdId := binary.LittleEndian.Uint16(data)
	offset += 2
	m := New()
	m.Type = Request
	m.ID = 1
	route := strconv.Itoa(int(cmdId))
	m.Route = route
	m.Data = data[4:]
	return m, nil
}

func DecodeData(data []byte) {
	cmdId := binary.LittleEndian.Uint32(data)
	fmt.Println("cmd=", cmdId)
}
