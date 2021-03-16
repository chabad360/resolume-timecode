package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/chabad360/go-osc/osc"
	"github.com/go-playground/pure/v5"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	broadcast = &Distributor{
		l: map[string]chan []byte{},
	}
	OSCOutPort string = "7001"
	OSCPort    string = "7000"
	OSCAddr    string = "127.0.0.1"
	httpPort   string = "8080"
	clipPath   string = "/composition/selectedclip"

	//go:embed index.html
	//go:embed main.js
	fs embed.FS

	p          = pure.New()
	httpServer *http.Server
	conn       net.PacketConn
	wg         sync.WaitGroup
	running    bool
	message    = &osc.Message{Arguments: []interface{}{"?"}}
	client     = &osc.Client{IP: OSCPort, Port: 7000}
	msg        string
)

func main() {
	//pr := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	//defer pr.Stop()

	p.Get("/ws", websocketStart)
	p.Get("/", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	p.Get("/main.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)

	gui()

	serverStop()
}

func serverStart() error {
	var err error
	if running {
		return nil
	}

	port, _ := strconv.Atoi(OSCPort)
	client.IP = OSCAddr
	client.Port = port

	conn, err = net.ListenPacket("udp", ":"+OSCOutPort)
	if err != nil {
		return fmt.Errorf("couldn't listen: %w", err)
	}

	wg.Add(1)
	go listenOSC(conn, &wg)

	wg.Add(1)
	go func() {
		for !running {
		}
		for running {
			time.Sleep(time.Second)
			message.Address = fmt.Sprintf("%s/name", clipPath)
			client.Send(message)
		}
		wg.Done()
	}()

	httpServer = &http.Server{Addr: ":" + httpPort, Handler: p.Serve()}

	wg.Add(1)
	go func() {
		httpServer.ListenAndServe()
		wg.Done()
	}()

	running = true
	return nil
}

func serverStop() {
	broadcast.Publish([]byte("/stop "))
	ctx, c := context.WithTimeout(context.Background(), time.Second*3)
	err := httpServer.Shutdown(ctx)
	if conn != nil {
		conn.Close()
	}
	if err != nil {
		httpServer.Close()
	}
	c()
	running = false
	wg.Wait()
}

func getIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP
}

func listenOSC(conn net.PacketConn, wg *sync.WaitGroup) {
	defer wg.Done()
	server := &osc.Server{}
	for {
		packet, err := server.ReceivePacket(conn)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
		}

		if packet != nil {
			switch p := packet.(type) {
			default:
				continue
			case *osc.Message:
				msg = p.String()
				if strings.Contains(msg, clipPath) {
					broadcast.Publish([]byte(msg[len(clipPath):]))
				}

			case *osc.Bundle:
				for _, message := range p.Messages {
					msg = message.String()
					if strings.Contains(msg, clipPath) {
						broadcast.Publish([]byte(msg[len(clipPath):]))
					}
				}
			}
		}
	}
}

func websocketStart(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	ctx := c.CloseRead(context.Background())

	l := broadcast.Listen(r.RemoteAddr)
	defer broadcast.Close(r.RemoteAddr)

	for {
		select {
		case <-ctx.Done():
			c.Close(websocket.StatusNormalClosure, "")
			return
		case m := <-l:
			err = c.Write(ctx, websocket.MessageText, m)
			if err != nil {
				//log.Println(err)
				return
			}
		}
	}
}
