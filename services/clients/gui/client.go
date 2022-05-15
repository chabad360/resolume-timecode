package gui

import (
	"context"
	"fyne.io/fyne/v2/data/binding"
	"resolume-timecode/services/clients"
	"resolume-timecode/util"
	"time"
)

var (
	m *util.Message
	t = time.NewTicker(time.Millisecond * 100)
)

func Init() {
	t.Stop()
	clients.Register("gui", func(message *util.Message) []byte {
		m = message
		return []byte{}
	})
}

func Start(c context.Context, start func(), done func()) {
	t.Reset(time.Millisecond * 110)

	go func() {
		start()
		defer done()
		for {
			select {
			case <-t.C:
				if m != nil {
					ClipNameBinding.Set("Clip Name: " + m.ClipName)
					TimeLeftBinding.Set(m.Hour + ":" + m.Minute + ":" + m.Second + "." + m.MS)
					ClipLengthBinding.Set("Clip Length: " + m.ClipLength)
				}
			case <-c.Done():
				t.Stop()
				ClipNameBinding.Set("Clip Name: None")
				TimeLeftBinding.Set("00:00:00.00")
				ClipLengthBinding.Set("Clip Length: 0.000s")
				return
			}
		}
	}()
}

var (
	ClipLengthBinding = binding.NewString()
	TimeLeftBinding   = binding.NewString()
	ClipNameBinding   = binding.NewString()
)
