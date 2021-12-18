package message

import (
	"github.com/topfreegames/pitaya/conn/codec"
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

	return nil, nil
}

// Decode decodes the message
func (pme *PapertigerMessageEncoder) Decode(data []byte) (*Message, error) {
	if len(data) < 4 {
		return nil, ErrInvalidMessage
	}
	cmdId := codec.BytesToInt(data[0:4])
	m := New()
	m.Type = Request
	m.ID = 0
	route := strconv.Itoa(cmdId)
	m.Route = route
	m.Data = data[4:]
	return m, nil
}
