package util

import (
	"strings"
	"sync"
)

type Distributor struct {
	l map[string]chan []byte
	m sync.RWMutex

	e map[string]func(*Message) []byte
}

func NewDistributor() *Distributor {
	return &Distributor{
		l: make(map[string]chan []byte),
		e: make(map[string]func(*Message) []byte),
	}
}

func (d *Distributor) Register(name string, f func(*Message) []byte) {
	d.e[name] = f
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

func (d *Distributor) Publish(m *Message) {
	e := make(map[string][]byte)
	for k, v := range d.e {
		e[k] = v(m)
	}
	d.publish(e)
}

func (d *Distributor) publish(e map[string][]byte) {
	d.m.RLock()

	for k, ch := range d.l {
		v := e[strings.Split(k, "/")[0]]
		select {
		case ch <- v:
		default:
		}
	}
	d.m.RUnlock()
}
