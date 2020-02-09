package yar

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/ruolinn/micro-yar/codec/yarrpc"

	"github.com/micro/go-micro/transport"
)

type yarTransportSocket struct {
	yt   *yarTransport
	conn net.Conn
	once sync.Once
	sync.Mutex
}

func (y *yarTransportSocket) Recv(m *transport.Message) error {
	header := yarrpc.YarHeader{}
	header.Reset()

	binary.Read(y.conn, binary.BigEndian, &header.Id)
	binary.Read(y.conn, binary.BigEndian, &header.Version)
	binary.Read(y.conn, binary.BigEndian, &header.MagicNum)
	binary.Read(y.conn, binary.BigEndian, &header.Reserved)
	binary.Read(y.conn, binary.BigEndian, &header.Provider)
	binary.Read(y.conn, binary.BigEndian, &header.Token)
	binary.Read(y.conn, binary.BigEndian, &header.BodyLen)
	binary.Read(y.conn, binary.BigEndian, &header.Packager)

	if header.BodyLen < 8 || header.BodyLen > 2*1024*1024 {
		return errors.New("yar: Response header missing params")
	}

	body_len := (int)(header.BodyLen - 8)
	data := make([]byte, body_len)
	n, err := io.ReadFull(y.conn, data)
	if n != body_len {
		return errors.New("yar: readPack body len error" + err.Error())
	}

	if m.Header == nil {
		m.Header = make(map[string]string)
	}

	m.Header["Content-Type"] = "application/yar"
	m.Header["Packager"] = "msgpack"
	m.Body = data

	return nil
}

func (y *yarTransportSocket) Close() error {
	err := y.conn.Close()
	return err
}

func (y *yarTransportSocket) Send(m *transport.Message) error {
	_, err := y.conn.Write(m.Body)

	return err
}

func (y *yarTransportSocket) Local() string {
	return ""
}

func (y *yarTransportSocket) Remote() string {
	return ""
}
