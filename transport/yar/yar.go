package yar

import (
	"net"

	"github.com/micro/go-micro/transport"
	mnet "github.com/micro/misc/lib/net"
)

type yarTransport struct {
	opts transport.Options
}

func NewYarTransport(opts ...transport.Option) *yarTransport {
	var options transport.Options
	for _, o := range opts {
		o(&options)
	}

	return &yarTransport{opts: options}
}

func (y *yarTransport) Listen(addr string, opts ...transport.ListenOption) (transport.Listener, error) {
	var options transport.ListenOptions
	for _, o := range opts {
		o(&options)
	}

	var l net.Listener
	var err error

	fn := func(addr string) (net.Listener, error) {
		return net.Listen("tcp", addr)
	}

	l, err = mnet.Listen(addr, fn)

	if err != nil {
		return nil, err
	}

	return &yarTransportListener{
		yt:       y,
		listener: l,
	}, nil
}

func (y *yarTransport) Dial(addr string, opts ...transport.DialOption) (transport.Client, error) {
	return nil, nil
}

func (y *yarTransport) String() string {
	return "yar"
}

func (y *yarTransport) Init(...transport.Option) error {
	return nil
}

func (y *yarTransport) Options() transport.Options {
	return y.opts
}
