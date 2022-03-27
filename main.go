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
	clipPath      = a.Preferences().StringWithFallback("clipPath", "")
	clientMessage = ""
	clipInvert    = a.Preferences().BoolWithFallback("clipInvert", false)

	//go:embed index.html
	//go:embed main.js
	//go:embed images/favicon.png
	//go:embed osc.min.js
	//go:embed osc.min.js.map
	fs embed.FS

	m          = http.NewServeMux()
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

	m.HandleFunc("/", websocketStart)
	//m.HandleFunc("/", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/main.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/osc.min.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/osc.min.js.map", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/images/favicon.png", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)

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

	httpServer = &http.Server{Addr: ":" + httpPort, Handler: m}

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
		broadcast.Publish(osc.NewMessage("/stop"))
		broadcast.Send()
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
	conn, err := net.Dial("udp", "8.8.8.8:80") //TODO: fix this
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
			procMsg(data)

		case *osc.Bundle:
			for _, elem := range data.Elements {
				handleOSC(elem, a)
			}
		}
	}
}

func websocketStart(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") == "" {
		http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP(w, r)
		return
	}
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	ctx := c.CloseRead(context.Background())

	//m, _ := osc.NewMessage("/open").MarshalBinary()
	//if c.Write(ctx, websocket.MessageBinary, m) != nil {
	//	return
	//}

	b, _ := osc.NewBundle(osc.NewMessage("/message", clientMessage), osc.NewMessage("/name", clipName), osc.NewMessage("/tminus", !clipInvert)).MarshalBinary()
	if c.Write(ctx, websocket.MessageBinary, b) != nil {
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
			if c.Write(ctx, websocket.MessageBinary, m) != nil {
				//log.Println(err)
				return
			}
		}
	}
}
