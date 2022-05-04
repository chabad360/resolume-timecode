package main

import (
	"fyne.io/fyne/v2/app"
	"resolume-timecode/gui"
)

var (
	//	broadcast = &util.Distributor{
	//		l: map[string]chan []byte{},
	//	}
	a = app.NewWithID("me.chabad360.resolume-timecode")

	OSCOutPort    = a.Preferences().StringWithFallback("OSCOutPort", "7001")
	OSCPort       = a.Preferences().StringWithFallback("OSCPort", "7000")
	OSCAddr       = a.Preferences().StringWithFallback("OSCAddr", "127.0.0.1")
	httpPort      = a.Preferences().StringWithFallback("httpPort", "8080")
	clipPath      = a.Preferences().StringWithFallback("clipPath", "")
	clientMessage = ""
	clipInvert    = a.Preferences().BoolWithFallback("clipInvert", false)
	//
	//	//go:embed services/clients/http/index.html
	//	//go:embed services/clients/http/main.js
	//	//go:embed images/favicon.png
	//	//go:embed services/clients/http/osc.min.js
	//	//go:embed services/clients/http/osc.min.js.map
	//	fs embed.FS
	//
	//	m          = http.NewServeMux()
	//	httpServer *http.Server
	//	oscServer  *osc.Server
	//	wg         sync.WaitGroup
	//	running    bool
	//	message    = &osc.Message{Arguments: []interface{}{"?"}}
	//	message2   = &osc.Message{Arguments: []interface{}{"?"}}
	//	t          = time.Tick(time.Minute)
)

func main() {
	//pr := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	//defer pr.Stop()

	gui.gui()
}
