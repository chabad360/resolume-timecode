package clients

import (
	"context"
	"github.com/chabad360/go-osc/osc"
	"resolume-timecode/util"
)

var (
	broadcast = util.NewDistributor()
)

type Client interface {
	Start(context.Context) error
	New() error
	SetAddress(string, string) error
}

func Register(key string) <-chan []byte {
	return broadcast.Listen(key)
}

func Close(key string) {
	broadcast.Close(key)
}

func PublishMultiple(m ...osc.Packet) {
	broadcast.PublishMultipleAndSend(m...)
}

func Publish(m *osc.Message) {
	broadcast.Publish(m)
}

func Send() {
	broadcast.Send()
}
