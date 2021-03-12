package main

import "sync"

type Distributor struct {
	l map[string]chan interface{}
	m sync.RWMutex
}

func New() *Distributor {
	return &Distributor{
		l: map[string]chan interface{}{},
	}
}

func (d *Distributor) Listen(key string) <-chan interface{} {
	ch := make(chan interface{})
	d.m.Lock()
	defer d.m.Unlock()

	if och, ok := d.l[key]; ok {
		close(och)
	}

	d.l[key] = ch
	return ch
}

func (d *Distributor) Close(key string) {
	d.m.Lock()
	defer d.m.Unlock()

	if och, ok := d.l[key]; ok {
		close(och)
	}

	delete(d.l, key)
}

func (d *Distributor) Publish(v interface{}) {
	d.m.RLock()
	defer d.m.RUnlock()

	for _, ch := range d.l {
		select {
		case ch <- v:
		default:
		}
	}
}
