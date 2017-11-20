package yarrpc

import (
	"errors"

	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

type msgpackPack struct {
	Name [8]byte
}

func NewMsgpack() *msgpackPack {
	name := [8]byte{'M', 'S', 'G', 'P', 'A', 'C', 'K'}

	return &msgpackPack{
		Name: name,
	}
}

func (m *msgpackPack) Marshal(x interface{}) (data []byte, err error) {
	if x == nil {
		return nil, errors.New("yar: serverResponse null")
	}
	return msgpack.Marshal(x)
}

func (m *msgpackPack) GetName() [8]byte {
	return m.Name
}

func (m *msgpackPack) Unmarshal(data []byte, x interface{}) error {
	return msgpack.Unmarshal(data, &x)
}
