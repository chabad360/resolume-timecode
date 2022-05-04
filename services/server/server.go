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

//func serverStart() error {
//	if running {
//		return nil
//	}
//
//	oscServer = &osc.Server{Addr: ":" + OSCOutPort, Handler: handleOSC}
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		oscServer.ListenAndServe()
//	}()
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for !running {
//		}
//		for running {
//			select {
//			case <-t:
//				runtime.GC()
//			default:
//				time.Sleep(time.Millisecond * 100)
//			}
//		}
//	}()
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for !running {
//		}
//		for running {
//			time.Sleep(time.Millisecond * 110)
//			gui.clipLengthBinding.Set(fmt.Sprintf("Clip Length: %.3fs", server.clipLength))
//			gui.timeLeftBinding.Set(server.timeLeft)
//		}
//		gui.timeLeftBinding.Set("00:00:00.000")
//		gui.clipLengthBinding.Set("Clip Length: 0.000s")
//	}()
//
//	running = true
//	return nil
//}
//
//func serverStop() {
//	if running {
//		broadcast.Publish(osc.NewMessage("/stop"))
//		broadcast.Send()
//		ctx, c := context.WithTimeout(context.Background(), time.Second*3)
//		err := httpServer.Shutdown(ctx)
//		oscServer.Close()
//		if err != nil {
//			httpServer.Close()
//		}
//		c()
//		running = false
//		wg.Wait()
//	}
//}

func Start(c context.Context, start func(), done func()) error {
	oscServer = &osc.Server{Addr: ":" + config.GetString(config.OSCOutPort), Handler: handleOSC}

	go func() {
		start()
		defer done()
		oscServer.ListenAndServe()
	}()

	go func() {
		select {
		case <-c.Done():
			oscServer.Close()
		}
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
