package yarrpc

import (
	"encoding/binary"
	"errors"
	"io"
	"strconv"
	"strings"
	"sync"

	msgpack "gopkg.in/vmihailenco/msgpack.v2"

	"github.com/micro/go-micro/codec"
)

const YAR_PROTOCOL_MAGIC_NUM = 0x80DFEC60
const YAR_PROTOCOL_VERSION = 0
const YAR_PROTOCOL_RESERVED = 0

var YAR_PROTOCOL_TOKEN = [32]byte{}
var YAR_PROVIDER = [32]byte{'Y', 'a', 'r', ' ', 'G', 'o', ' ', 'C', 'l', 'i', 'e', 'n', 't'}

type Packager interface {
	Unmarshal(data []byte, x interface{}) error
	Marshal(interface{}) ([]byte, error)
	GetName() [8]byte
}

var packagers map[string]Packager

func init() {
	packagers = make(map[string]Packager)
	msgpack := NewMsgpack()

	packagers["msgp"] = msgpack
}

type serverCodec struct {
	//rwc    io.ReadWriteCloser
	dec    *msgpack.Decoder
	enc    *msgpack.Encoder
	c      io.ReadWriteCloser
	req    serverRequest
	packer Packager
	sync.Mutex
	// 数据类型更新
	seq     uint64
	pending map[uint64]int64
}
type RawMessage []byte

type serverRequest struct {
	Header *YarHeader  `json:"_" msgpack:"_"`
	Id     int64       `json:"i" msgpack:"i"`
	Method string      `json:"m" msgpack:"m"`
	Params *RawMessage `json:"p" msgpack:"p"`
}

func (m *RawMessage) MarshalJSON() ([]byte, error) {
	return *m, nil
}

func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

func (m *RawMessage) MarshalMsgpack() ([]byte, error) {
	return *m, nil
}

func (m *RawMessage) UnmarshalMsgpack(data []byte) error {
	if m == nil {
		return errors.New("msgpack.RawMessage: UnmarshalMsgpack on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

type serverResponse struct {
	Id     int64             `json:"i" msgpack:"i"`
	Error  map[string]string `json:"e" msgpack:"e"`
	Output string            `json:"o" msgpack:"o"`
	Status int               `json:"s" msgpack:"s"`
	Result *interface{}      `json:"r" msgpack:"r"`
}

type YarHeader struct {
	Id       int32  // transaction id
	Version  uint16 // jsoncl version
	MagicNum uint32 // default is: 0x80DFEC60
	Reserved uint32
	Provider [32]byte // reqeust from who
	Token    [32]byte // request token, used for authentication
	BodyLen  uint32   // request body len
	Packager [8]byte  // packager
}

func (h *YarHeader) Reset() {
	h.Id = 0
	h.Version = 0
	h.MagicNum = 0
	h.Reserved = 0
	h.BodyLen = 0
}

func newServerCodec(conn io.ReadWriteCloser) *serverCodec {
	return &serverCodec{
		dec:     msgpack.NewDecoder(conn),
		enc:     msgpack.NewEncoder(conn),
		c:       conn,
		pending: make(map[uint64]int64),
	}
}

func (r *serverRequest) reset() {
	r.Method = ""
	r.Params = nil
	r.Id = 0
	//r.Header = nil
}

func (c *serverCodec) ReadHeader(m *codec.Message) error {

	if err := c.dec.Decode(&c.req); err != nil {
		return err
	}

	m.Method = c.req.Method
	c.packer, _ = getPackagerBybyte(m.Header["Packager"])

	c.Lock()
	c.seq++
	c.pending[c.seq] = c.req.Id
	c.req.Id = 0
	m.Id = strconv.FormatInt(int64(c.seq), 10)
	c.Unlock()

	return nil
}

func (c *serverCodec) ReadBody(x interface{}) error {
	if x == nil {
		return nil
	}

	return c.packer.Unmarshal(*c.req.Params, &x)
}

func (c *serverCodec) Write(m *codec.Message, x interface{}) error {
	var resp serverResponse

	c.Lock()
	idx, _ := strconv.ParseInt(m.Id, 10, 64)
	//更新数据类型
	b, ok := c.pending[uint64(idx)]
	if !ok {
		c.Unlock()
		return errors.New("invalid sequence number in response")
	}
	c.Unlock()

	if b == 0 {
		// Invalid request so no id.  Use JSON null.
		b = 0 //&null
	}

	resp = serverResponse{
		Id:     b,
		Status: 0x0,
	}
	resp.Result = &x

	data, err := msgpack.Marshal(&resp)
	if err != nil {
		return err
	}

	header := YarHeader{}
	header.Reset()

	header.Id = (int32)(b)
	header.Version = YAR_PROTOCOL_RESERVED
	header.MagicNum = YAR_PROTOCOL_MAGIC_NUM
	header.Reserved = YAR_PROTOCOL_RESERVED
	header.Provider = YAR_PROVIDER
	header.Token = YAR_PROTOCOL_TOKEN
	header.BodyLen = uint32(len(data) + 8)
	header.Packager = [8]byte{'m', 's', 'g', 'p', 'a', 'c', 'k'}

	binary.Write(c.c, binary.BigEndian, header)

	_, err = c.c.Write(data)

	if err != nil {
		return err
	}

	return nil
}

func getPackagerBybyte(name string) (Packager, error) {
	var buf [8]byte
	copy(buf[:], name)

	return getPackager(strings.ToLower(string(buf[:4])))
}

//测试用
func GetPack(name string) (Packager, error) {
	return getPackagerBybyte(name)
}

func getPackager(name string) (Packager, error) {
	packer, ok := packagers[name]
	if !ok {
		return nil, errors.New("not packager")
	}

	return packer, nil
}
