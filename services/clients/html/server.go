package html

import (
	"context"
	"embed"
	"github.com/francoispqt/gojay"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"resolume-timecode/config"
	"resolume-timecode/services/clients"
	"resolume-timecode/util"
)

var (
	//go:embed index.html
	//go:embed main.js
	fs embed.FS

	Icon []byte
)

type Server struct {
	c          context.Context
	m          *http.ServeMux
	httpServer *http.Server
	l          net.Listener
}

func (s *Server) Start(c context.Context, start func(), done func()) error {
	s.c = c
	var err error

	s.httpServer = &http.Server{Handler: s.m}
	s.httpServer.BaseContext = func(_ net.Listener) context.Context {
		return s.c
	}

	listenConfig := &net.ListenConfig{}
	s.l, err = listenConfig.Listen(s.c, "tcp", ":"+config.GetString(config.HTTPPort))
	if err != nil {
		return err
	}

	go func() {
		start()
		defer done()
		s.httpServer.Serve(s.l)
	}()

	go func() {
		<-c.Done()
		s.stop()
	}()

	return nil
}

func (s *Server) stop() {
	err := s.httpServer.Shutdown(context.Background())
	if err != nil {
		s.httpServer.Close()
	}
}

func New() *Server {
	clients.Register("html", func(m *util.Message) []byte {
		b, _ := gojay.Marshal(m)
		return b
	})

	m := http.NewServeMux()
	m.HandleFunc("/", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/ws", websocketStart)
	m.HandleFunc("/main.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/images/favicon.png", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "image/png")
		writer.Write(Icon)
	})

	return &Server{m: m}
}

func websocketStart(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	ctx := c.CloseRead(r.Context())

	l := clients.Listen("html/" + r.RemoteAddr)
	defer clients.Close("html/" + r.RemoteAddr)

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
