package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/go-playground/pure/v5"
	"github.com/hypebeast/go-osc/osc"
	"log"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	Server            = server{}
	broadcast         = New()
	OSCOutPort string = "7001"
	OSCPort    string = "7000"
	OSCAddr    string = "127.0.0.1"
	httpPort   string = "8080"
	clipPath   string = "/composition/selectedclip"

	//go:embed index.html
	//go:embed main.js
	fs embed.FS
)

type server struct {
	httpServer *http.Server
	conn       net.PacketConn
	wg         sync.WaitGroup
	running    bool
}

func main() {
	gui()

	Server.Stop()
}

func (s *server) Start() error {
	var err error
	if s.running {
		return nil
	}

	port, _ := strconv.Atoi(OSCPort)
	client := osc.NewClient(OSCAddr, port)

	s.conn, err = net.ListenPacket("udp", ":"+OSCOutPort)
	if err != nil {
		return fmt.Errorf("Couldn't listen: %w", err)
	}

	s.wg.Add(1)
	go listenOSC(s.conn, &s.wg)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for !s.running {
		}
		for s.running {
			time.Sleep(time.Second)
			message := osc.NewMessage(clipPath + "/name")
			message.Append("?")
			client.Send(message)
		}
	}()

	p := pure.New()

	p.Get("/ws", websocketStart)
	p.Get("/", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	p.Get("/main.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)

	s.httpServer = &http.Server{Addr: ":" + httpPort, Handler: p.Serve()}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		}
	}()

	s.running = true
	return nil
}

func (s *server) Stop() {
	broadcast.Publish("/stop ")
	ctx, c := context.WithTimeout(context.Background(), time.Second*3)
	err := s.httpServer.Shutdown(ctx)
	s.conn.Close()
	if err != nil {
		s.httpServer.Close()
	}
	c()
	s.running = false
	s.wg.Wait()
}

func getIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
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
			switch packet.(type) {
			default:

			case *osc.Message:
				msg := packet.(*osc.Message).String()
				if strings.Contains(msg, clipPath) {
					broadcast.Publish(packet.(*osc.Message).String())
				}

			case *osc.Bundle:
				bundle := packet.(*osc.Bundle)
				for _, message := range bundle.Messages {
					msg := message.String()
					if strings.Contains(msg, clipPath) {
						broadcast.Publish(message.String())
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
			msg := m.(string)
			err = c.Write(ctx, websocket.MessageText, []byte(msg))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
