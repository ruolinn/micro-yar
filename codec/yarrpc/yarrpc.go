package yarrpc

import (
	"fmt"
	"io"

	"github.com/micro/go-micro/codec"
)

type yarCodec struct {
	mt  codec.MessageType
	rwc io.ReadWriteCloser
	s   *serverCodec
}

func (y *yarCodec) ReadHeader(m *codec.Message, mt codec.MessageType) error {
	y.mt = mt

	switch mt {
	case codec.Request:
		return y.s.ReadHeader(m)
	}

	return nil
}

func (y *yarCodec) ReadBody(b interface{}) error {
	switch y.mt {
	case codec.Request:
		return y.s.ReadBody(b)
	}
	return nil
}

func (y *yarCodec) String() string {
	return "yar-rpc"
}

func (y *yarCodec) Write(m *codec.Message, b interface{}) error {
	switch m.Type {
	case codec.Response:
		return y.s.Write(m, b)
	default:
		return fmt.Errorf("Unrecognised message type: %v", m.Type)
	}

}

func NewCodec(rwc io.ReadWriteCloser) codec.Codec {
	return &yarCodec{
		rwc: rwc,
		s:   newServerCodec(rwc),
	}
}

func (y *yarCodec) Close() error {
	return nil
}
