package util

import (
	"github.com/chabad360/go-osc/osc"
	"sync"
)

type Distributor struct {
	l  map[string]chan []byte
	m  sync.RWMutex
	b  *osc.Bundle
	bM sync.Mutex
}

func (d *Distributor) Listen(key string) <-chan []byte {
	ch := make(chan []byte)
	d.m.Lock()

	if och, ok := d.l[key]; ok {
		close(och)
	}

	d.l[key] = ch
	d.m.Unlock()
	return ch
}

func (d *Distributor) Close(key string) {
	d.m.Lock()

	if och, ok := d.l[key]; ok {
		close(och)
	}

	delete(d.l, key)
	d.m.Unlock()
}

func (d *Distributor) Publish(m *osc.Message) {
	d.bM.Lock()

	if d.b == nil {
		d.b = &osc.Bundle{Timetag: osc.NewImmediateTimetag()}
	}

	d.b.Elements = append(d.b.Elements, m)

	d.bM.Unlock()
}

func (d *Distributor) Send() {
	d.bM.Lock()

	b, err := d.b.MarshalBinary()
	if err != nil {
		panic(err)
	}

	d.b = nil

	d.publish(b)
	d.bM.Unlock()
}

func (d *Distributor) publish(v []byte) {
	d.m.RLock()

	for _, ch := range d.l {
		select {
		case ch <- v:
		default:
		}
	}
	d.m.RUnlock()
}
