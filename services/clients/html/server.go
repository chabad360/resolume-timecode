package html

import (
	"context"
	"embed"
	"github.com/chabad360/go-osc/osc"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"resolume-timecode/config"
	"resolume-timecode/services/clients"
)

var (
	////go:embed images/favicon.png

	//go:embed index.html
	//go:embed main.js
	//go:embed osc.min.js
	//go:embed osc.min.js.map
	fs embed.FS
)

type Server struct {
	c          context.Context
	m          *http.ServeMux
	httpServer *http.Server
}

func (s *Server) Start(c context.Context, start func(), done func()) error {
	s.c = c

	s.httpServer = &http.Server{Addr: ":" + config.GetString(config.HTTPPort), Handler: s.m}
	s.httpServer.BaseContext = func(_ net.Listener) context.Context {
		return s.c
	}

	go func() {
		start()
		defer done()
		s.httpServer.ListenAndServe()
	}()

	go func() {
		<-c.Done()
		s.stop()
		return
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
	m := http.NewServeMux()
	m.HandleFunc("/", websocketStart)
	//m.HandleFunc("/", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/main.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/osc.min.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/osc.min.js.map", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	//m.HandleFunc("/images/favicon.png", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)

	return &Server{m: m}
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

	b, _ := osc.NewBundle(osc.NewMessage("/message", config.GetString(config.ClientMessage))).MarshalBinary()
	//b, _ := osc.NewBundle(osc.NewMessage("/message", config.GetString(config.ClientMessage)), osc.NewMessage("/name", server.clipName), osc.NewMessage("/tminus", !clipInvert)).MarshalBinary()
	if c.Write(ctx, websocket.MessageBinary, b) != nil {
		return
	}

	l := clients.Register(r.RemoteAddr)
	defer clients.Close(r.RemoteAddr)

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
