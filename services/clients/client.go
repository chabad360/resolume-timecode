package clients

import (
	"resolume-timecode/util"
)

var (
	broadcast = util.NewDistributor()
)

func Register(key string, e func(*util.Message) []byte) {
	broadcast.Register(key, e)
}

func Listen(key string) <-chan []byte {
	return broadcast.Listen(key)
}

func Close(key string) {
	broadcast.Close(key)
}

func Publish(m *util.Message) {
	broadcast.Publish(m)
}
