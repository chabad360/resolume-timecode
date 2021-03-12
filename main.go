package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/chabad360/multicast"
	"github.com/go-playground/pure/v5"
	mw "github.com/go-playground/pure/v5/_examples/middleware/logging-recovery"
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
	broadcast         = multicast.New()
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
	ticker     *time.Ticker
	running    bool
}

func main() {
	gui()

	Server.Stop()
}

func (s *server) Start() {
	if s.running {
		return
	}

	port, err := strconv.Atoi(OSCPort)
	if err != nil {
		log.Fatal(err)
	}
	client := osc.NewClient(OSCAddr, port)

	s.conn, err = net.ListenPacket("udp", ":"+OSCOutPort)
	if err != nil {
		fmt.Println("Couldn't listen: ", err)
	}

	s.wg.Add(1)
	go listenOSC(s.conn, &s.wg)

	if s.ticker == nil {
		s.ticker = time.NewTicker(time.Second)
	} else {
		s.ticker.Reset(time.Second)
	}

	go func() {
		for _ = range s.ticker.C {
			message := osc.NewMessage(clipPath + "/name")
			message.Append("?")
			client.Send(message)
		}
	}()

	p := pure.New()
	p.Use(mw.LoggingAndRecovery(true))

	p.Get("/ws", websocketStart)
	p.Get("/", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	p.Get("/main.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)

	s.httpServer = &http.Server{Addr: ":" + httpPort, Handler: p.Serve()}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	s.running = true
}

func (s *server) Stop() {
	broadcast.C <- "/stop "
	s.ticker.Stop()
	ctx, c := context.WithTimeout(context.Background(), time.Second*3)
	err := s.httpServer.Shutdown(ctx)
	s.conn.Close()
	if err != nil {
		s.httpServer.Close()
	}
	c()
	s.wg.Wait()
	s.running = false
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
			fmt.Println("Server error: " + err.Error())
		}

		if packet != nil {
			switch packet.(type) {
			default:
				fmt.Println("Unknown packet type!")

			case *osc.Message:
				msg := packet.(*osc.Message).String()
				if strings.Contains(msg, clipPath) {
					broadcast.C <- packet.(*osc.Message).String()
				}

			case *osc.Bundle:
				bundle := packet.(*osc.Bundle)
				for _, message := range bundle.Messages {
					msg := message.String()
					if strings.Contains(msg, clipPath) {
						broadcast.C <- message.String()
					}
				}
			}
		}
	}
}

func websocketStart(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	ctx := c.CloseRead(context.Background())

	l := broadcast.Listen()

	for {
		select {
		case <-ctx.Done():
			c.Close(websocket.StatusNormalClosure, "")
			return
		case m := <-l.C:
			msg := m.(string)
			err = c.Write(ctx, websocket.MessageText, []byte(msg))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
