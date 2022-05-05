package osc

import (
	"context"
	"errors"
	"github.com/chabad360/go-osc/osc"
	"net"
	"resolume-timecode/config"
	"resolume-timecode/services/clients"
	"resolume-timecode/util"
)

type Client struct {
	conn net.Conn
}

func New() (*Client, error) {
	clients.Register("osc", func(m *util.Message) []byte {
		b, _ := osc.NewBundle(
			osc.NewMessage("/hour", m.Hour),
			osc.NewMessage("/minute", m.Minute),
			osc.NewMessage("/second", m.Second),
			osc.NewMessage("/ms", m.MS),
			osc.NewMessage("/length", m.ClipLength),
			osc.NewMessage("/name", m.ClipName),
			osc.NewMessage("/tminus", !m.Invert),
			osc.NewMessage("/message", m.Message),
		).MarshalBinary()
		return b
	})

	conn, err := net.Dial("udp", config.GetString(config.OSCClientAddr)+":"+config.GetString(config.OSCClientPort))
	if err != nil {
		return nil, err
	}

	return &Client{conn}, nil
}

func (c *Client) Start(ctx context.Context, start func(), done func()) error {
	if c.conn == nil {
		return errors.New("OSC client not initialized")
	}

	go func() {
		start()
		defer done()

		l := clients.Listen("osc/" + c.conn.RemoteAddr().String())
		defer clients.Close("osc/" + c.conn.RemoteAddr().String())

		for {
			select {
			case <-ctx.Done():
				c.conn.Close()
				return
			case m := <-l:
				c.conn.Write(m)
			}
		}
	}()

	return nil
}