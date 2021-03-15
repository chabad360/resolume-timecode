package main

import "sync"

type Distributor struct {
	l map[string]chan string
	m sync.RWMutex
}

func New() *Distributor {
	return &Distributor{
		l: map[string]chan string{},
	}
}

func (d *Distributor) Listen(key string) <-chan string {
	ch := make(chan string)
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

func (d *Distributor) Publish(v string) {
	d.m.RLock()
	defer d.m.RUnlock()

	for _, ch := range d.l {
		select {
		case ch <- v:
		default:
		}
	}
}
