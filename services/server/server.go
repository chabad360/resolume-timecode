package server

import (
	"context"
	"github.com/chabad360/go-osc/osc"
	"net"
	"resolume-timecode/config"
)

var (
	oscServer *osc.Server
)

func Start(c context.Context, start func(), done func()) error {
	oscServer = &osc.Server{Addr: ":" + config.GetString(config.OSCOutPort), Handler: handleOSC}

	go func() {
		start()
		defer done()
		oscServer.ListenAndServe()
	}()

	go func() {
		<-c.Done()
		oscServer.Close()
	}()

	return nil
}

func handleOSC(packet osc.Packet, a net.Addr) {
	if packet != nil {
		switch data := packet.(type) {
		case *osc.Message:
			procMsg(data)

		case *osc.Bundle:
			for _, elem := range data.Elements {
				handleOSC(elem, a)
			}
		}
	}
}
