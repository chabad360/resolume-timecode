package main

//go:generate env GOOS=js GOARCH=wasm go build -o=main.wasm wasm/calculation.go wasm/main.go

import (
	"context"
	"embed"
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"log"
	"net"
	"net/http"
	"os"

	"time"

	"github.com/SierraSoftworks/multicast"
	"github.com/go-playground/pure/v5"
	mw "github.com/go-playground/pure/v5/_examples/middleware/logging-recovery"
	"nhooyr.io/websocket"
)

var (
	broadcast     = multicast.New()
	addr          = ":7001"
	listenOSCPort = 7000
	listenOSCAddr = "localhost"
	listenAddr    = ":80"

	//go:embed index.html
	//go:embed main.wasm
	fs embed.FS
)

func main() {
	client := osc.NewClient(listenOSCAddr, listenOSCPort)

	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Println("Couldn't listen: ", err)
	}
	defer conn.Close()

	go listenOSC(conn)

	t := time.NewTicker(time.Second)
	defer t.Stop()

	message := osc.NewMessage("/composition/selectedclip/name")
	message.Append("?")
	go func() {
		for {
			<-t.C
			client.Send(message)
		}
	}()

	p := pure.New()
	p.Use(mw.LoggingAndRecovery(true))

	p.Get("/ws", websocketStart)
	//p.Get("/config", config)
	p.Get("/", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	p.Get("/main.wasm", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)

	log.Fatal(http.ListenAndServe(listenAddr, p.Serve()))
}

func listenOSC(conn net.PacketConn) {
	server := &osc.Server{}
	for {
		packet, err := server.ReceivePacket(conn)
		if err != nil {
			fmt.Println("Server error: " + err.Error())
			os.Exit(1)
		}

		if packet != nil {
			switch packet.(type) {
			default:
				fmt.Println("Unknown packet type!")

			case *osc.Message:
				broadcast.C <- packet.(*osc.Message).String()

			case *osc.Bundle:
				bundle := packet.(*osc.Bundle)
				for _, message := range bundle.Messages {
					broadcast.C <- message.String()
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
