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
	oscServer = &osc.Server{Handler: handleOSC}

	listenConfig := &net.ListenConfig{}
	conn, err := listenConfig.ListenPacket(c, "udp", ":"+config.GetString(config.OSCOutPort))
	if err != nil {
		return err
	}

	go func() {
		start()
		defer done()
		oscServer.Serve(conn)
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
