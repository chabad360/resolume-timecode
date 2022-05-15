package services

import (
	"context"
	"resolume-timecode/config"
	"resolume-timecode/services/clients/gui"
	"resolume-timecode/services/clients/html"
	"resolume-timecode/services/clients/osc"
	"resolume-timecode/services/server"
	"sync"
)

var (
	wg     = sync.WaitGroup{}
	c      context.Context
	cancel context.CancelFunc

	running bool
)

func startReg() {
	wg.Add(1)
}

func done() {
	wg.Done()
}

func Start() error {
	if running {
		server.Reset()
		return nil
	}

	c, cancel = context.WithCancel(context.Background())

	if err := server.Start(c, startReg, done); err != nil {
		Stop()
		return err
	}

	gui.Init()
	gui.Start(c, startReg, done)

	if config.GetBool(config.EnableHttpClient) {
		if err := html.New().Start(c, startReg, done); err != nil {
			Stop()
			return err
		}
	}

	if config.GetBool(config.EnableOSCClient) {
		if err := osc.New().Start(c, startReg, done); err != nil {
			Stop()
			return err
		}
	}

	running = true
	return nil
}

func Stop() {
	running = false

	cancel()
	wg.Wait()
}
