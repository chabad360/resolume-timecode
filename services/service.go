package services

import (
	"context"
	"resolume-timecode/services/server"
	"sync"
)

var (
	wg     = sync.WaitGroup{}
	c      context.Context
	cancel context.CancelFunc
)

func start() {
	wg.Add(1)
}

func done() {
	wg.Done()
}

func Start() error {
	c, cancel = context.WithCancel(context.Background())
	server.Start(c, start, done)

	return nil
}

func Stop() {
	cancel()
	wg.Wait()
}
