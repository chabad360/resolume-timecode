package main

import (
	_ "embed"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"resolume-timecode/config"
	"resolume-timecode/gui"
	"resolume-timecode/services/clients/html"
)

var (
	a = app.NewWithID("me.chabad360.resolume-timecode")

	//go:embed images/logo.png
	logo []byte

	//go:embed images/favicon.png
	favicon []byte
)

func main() {
	//pr := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	//defer pr.Stop()

	config.Init(a)

	html.Icon = favicon

	gui.Gui(a, fyne.NewStaticResource("logo", logo))
}
