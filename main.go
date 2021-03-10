package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chabad360/multicast"
	"github.com/go-playground/pure/v5"
	mw "github.com/go-playground/pure/v5/_examples/middleware/logging-recovery"
	"nhooyr.io/websocket"
)

var (
	broadcast  = multicast.New()
	OSCOutPort int
	OSCPort    int
	OSCAddr    string
	httpPort   int
	clipPath   string

	//go:embed index.html
	//go:embed main.js
	fs embed.FS
)

func init() {
	flag.IntVar(&OSCPort, "osc-input-port", 7000, "Resolume OSC input port")
	flag.StringVar(&OSCAddr, "osc-addr", "localhost", "Address of device running Resolume")
	flag.IntVar(&OSCOutPort, "osc-output-port", 7001, "Port Resolume outputs OSC on (if you are using other OSC devices, make sure to set the broadcast address correctly)")
	flag.IntVar(&httpPort, "port", 8080, "Port that everyone uses to access this system (make sure it's open in your firewall)")
	flag.StringVar(&clipPath, "clip-path", "/composition/selectedclip", "OSC path for clip you want to track")
}

func main() {
	flag.Parse()

	client := osc.NewClient(OSCAddr, OSCPort)

	conn, err := net.ListenPacket("udp", ":"+strconv.Itoa(OSCOutPort))
	if err != nil {
		fmt.Println("Couldn't listen: ", err)
	}
	defer conn.Close()

	go listenOSC(conn)

	t := time.NewTicker(time.Second)
	defer t.Stop()

	message := osc.NewMessage(clipPath + "/name")
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
	p.Get("/main.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)

	p.Get("/path", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, clipPath)
	})

	fmt.Println("open your web browser to:", "http://"+getIP().String()+":"+strconv.Itoa(httpPort))

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(httpPort), p.Serve()))
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

func listenOSC(conn net.PacketConn) {
	server := &osc.Server{}
	for {
		packet, err := server.ReceivePacket(conn)
		if err != nil {
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
