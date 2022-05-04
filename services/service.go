package services

import (
	"context"
	"resolume-timecode/services/clients/gui"
	"resolume-timecode/services/clients/html"
	"resolume-timecode/services/server"
	"sync"
)

var (
	wg     = sync.WaitGroup{}
	c      context.Context
	cancel context.CancelFunc
)

func startReg() {
	wg.Add(1)
}

func done() {
	wg.Done()
}

func Start() error {
	c, cancel = context.WithCancel(context.Background())
	server.Start(c, startReg, done)
	html.New().Start(c, startReg, done)
	gui.Init()
	gui.Start(c, startReg, done)

	return nil
}

func Stop() {
	cancel()
	wg.Wait()
}
