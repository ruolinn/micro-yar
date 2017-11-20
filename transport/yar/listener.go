package yar

import (
	"net"
	"time"

	"github.com/go-log/log"
	"github.com/micro/go-micro/transport"
)

type yarTransportListener struct {
	yt       *yarTransport
	listener net.Listener
}

func (y *yarTransportListener) Accept(fn func(transport.Socket)) error {
	var tempDelay time.Duration

	for {
		c, err := y.listener.Accept()

		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Logf("http: Accept error: %v; retrying in %v\n", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		sock := &yarTransportSocket{
			yt:   y.yt,
			conn: c,
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Log("panic recovered: ", r)
				}
			}()

			fn(sock)
		}()
	}

}

func (y *yarTransportListener) Addr() string {
	return y.listener.Addr().String()
}

func (y *yarTransportListener) Close() error {
	return y.listener.Close()
}
