package main

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne/v2/app"
	"github.com/chabad360/go-osc/osc"
	"github.com/go-playground/pure/v5"
	"nhooyr.io/websocket"
)

var (
	broadcast = &Distributor{
		l: map[string]chan []byte{},
	}
	a = app.NewWithID("me.chabad360.resolume-timecode")

	OSCOutPort    = a.Preferences().StringWithFallback("OSCOutPort", "7001")
	OSCPort       = a.Preferences().StringWithFallback("OSCPort", "7000")
	OSCAddr       = a.Preferences().StringWithFallback("OSCAddr", "127.0.0.1")
	httpPort      = a.Preferences().StringWithFallback("httpPort", "8080")
	clipPath      = a.Preferences().StringWithFallback("clipPath", "/composition/selectedclip")
	clientMessage = ""

	//go:embed index.html
	//go:embed main.js
	//go:embed images/favicon.png
	fs embed.FS

	p          = pure.New()
	httpServer *http.Server
	oscServer  *osc.Server
	wg         sync.WaitGroup
	running    bool
	message    = &osc.Message{Arguments: []interface{}{"?"}}
	message2   = &osc.Message{Arguments: []interface{}{"?"}}
	t          = time.Tick(time.Minute)
)

func main() {
	//pr := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	//defer pr.Stop()

	p.Get("/ws", websocketStart)
	p.Get("/", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	p.Get("/main.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	p.Get("/images/favicon.png", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)

	gui()

	serverStop()
}

func serverStart() error {
	if running {
		return nil
	}

	oscServer = &osc.Server{Addr: ":" + OSCOutPort, Handler: handleOSC}

	wg.Add(1)
	go func() {
		defer wg.Done()
		oscServer.ListenAndServe()
	}()

	//message.Address = fmt.Sprintf("%s/name", clipPath)
	//if _, err := oscServer.WriteTo(message, OSCAddr+":"+OSCPort); err != nil {
	//	return err
	//}

	httpServer = &http.Server{Addr: ":" + httpPort, Handler: p.Serve()}

	wg.Add(1)
	go func() {
		defer wg.Done()
		httpServer.ListenAndServe()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for !running {
		}
		for running {
			select {
			case <-t:
				runtime.GC()
			default:
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for !running {
		}
		for running {
			time.Sleep(time.Millisecond * 110)
			clipLengthBinding.Set(fmt.Sprintf("Clip Length: %.3fs", clipLength))
			timeLeftBinding.Set(timeLeft)
		}
		timeLeftBinding.Set("00:00:00.000")
		clipLengthBinding.Set("Clip Length: 0.000s")
	}()

	running = true
	return nil
}

func serverStop() {
	if running {
		broadcast.Publish([]byte("/stop "))
		ctx, c := context.WithTimeout(context.Background(), time.Second*3)
		err := httpServer.Shutdown(ctx)
		oscServer.Close()
		if err != nil {
			httpServer.Close()
		}
		c()
		running = false
		wg.Wait()
	}
}

func getIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP
}

func handleOSC(packet osc.Packet, a net.Addr) {
	if packet != nil {
		switch data := packet.(type) {
		case *osc.Message:
			//fmt.Println(data)
			procMsg(data)

		case *osc.Bundle:
			for _, elem := range data.Elements {
				handleOSC(elem, a)
			}
		}
	}
}

func pushClientMessage() {
	broadcast.Publish([]byte("/message ,s " + clientMessage))
}

func websocketStart(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	ctx := c.CloseRead(context.Background())

	if c.Write(ctx, websocket.MessageText, []byte("/message ,s "+clientMessage)) != nil {
		return
	}
	if c.Write(ctx, websocket.MessageText, []byte("/name ,s "+clipName)) != nil {
		return
	}

	l := broadcast.Listen(r.RemoteAddr)
	defer broadcast.Close(r.RemoteAddr)

	for {
		select {
		case <-ctx.Done():
			c.Close(websocket.StatusNormalClosure, "")
			return
		case m := <-l:
			if c.Write(ctx, websocket.MessageText, m) != nil {
				//log.Println(err)
				return
			}
		}
	}
}
