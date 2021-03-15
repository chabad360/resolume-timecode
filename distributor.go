package main

import "sync"

type Distributor struct {
	l map[string]chan []byte
	m sync.RWMutex
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

func (d *Distributor) Publish(v []byte) {
	d.m.RLock()

	for _, ch := range d.l {
		select {
		case ch <- v:
		default:
		}
	}
	d.m.RUnlock()
}
