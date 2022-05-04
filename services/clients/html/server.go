package html

import (
	"context"
	"embed"
	"github.com/chabad360/go-osc/osc"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

var (
	//go:embed index.html
	//go:embed main.js
	//go:embed images/favicon.png
	//go:embed osc.min.js
	//go:embed osc.min.js.map
	fs embed.FS

	m          = http.NewServeMux()
	httpServer *http.Server
	wg         sync.WaitGroup
	running    bool
	message    = &osc.Message{Arguments: []interface{}{"?"}}
	message2   = &osc.Message{Arguments: []interface{}{"?"}}
	t          = time.Tick(time.Minute)
)

type Server struct {
	c context.Context
}

func (s *Server) Start(c context.Context) error {
	s.c = c

	httpServer = &http.Server{Addr: ":" + httpPort, Handler: m}

	wg.Add(1)
	go func() {
		defer wg.Done()
		httpServer.ListenAndServe()
	}()

}

func stop() {
	err := httpServer.Shutdown(ctx)
	if err != nil {
		httpServer.Close()
	}
}

func server() {
	m.HandleFunc("/", websocketStart)
	//m.HandleFunc("/", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/main.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/osc.min.js", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/osc.min.js.map", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)
	m.HandleFunc("/images/favicon.png", http.StripPrefix("/", http.FileServer(http.FS(fs))).ServeHTTP)

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

	b, _ := osc.NewBundle(osc.NewMessage("/message", clientMessage), osc.NewMessage("/name", server.clipName), osc.NewMessage("/tminus", !clipInvert)).MarshalBinary()
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
