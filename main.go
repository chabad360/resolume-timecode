package main

import (
	_ "embed"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"resolume-timecode/config"
	"resolume-timecode/gui"
)

var (
	a = app.NewWithID("me.chabad360.resolume-timecode")

	//go:embed images/logo.png
	logo []byte

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

	config.Init(a)

	gui.Gui(a, fyne.NewStaticResource("logo", logo))
}
