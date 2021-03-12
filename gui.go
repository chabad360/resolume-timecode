package main

import (
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
)

func gui() {
	a := app.New()
	w := a.NewWindow("Resolume Timecode sysServer")

	infoLabel := widget.NewLabel("sysServer Stopped")

	path := widget.NewEntry()
	path.SetText(clipPath)
	path.Validator = validation.NewRegexp(`^[^\?\,\[\]\{\}\#\ ]*$`, "not a valid OSC path")

	oscOutput := widget.NewEntry()
	oscOutput.SetText(OSCOutPort)
	oscOutput.Validator = validation.NewRegexp(`^[0-9]*$`, "not a valid port")

	oscInput := widget.NewEntry()
	oscInput.SetText(OSCPort)
	oscInput.Validator = validation.NewRegexp(`^[0-9]*$`, "not a valid port")

	oscAddr := widget.NewEntry()
	oscAddr.SetText(OSCAddr)
	oscAddr.Validator = validation.NewRegexp(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, "not a valid IP address")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Path", Widget: path, HintText: "OSC Path for clip to listen to"},
			{Text: "OSC Input Port", Widget: oscInput, HintText: "OSC Input port (usually 7000)"},
			{Text: "OSC Output Port", Widget: oscOutput, HintText: "OSC Output port (usually 7001) Note: If you have multiple services using Resolume OSC make use the correct broadcast address."},
			{Text: "OSC Host Address", Widget: oscAddr, HintText: "IP address of device that's running Resolume (make sure to open the OSC input port in your firewall)"},
		},
		SubmitText: "Start sysServer",
		CancelText: "Stop sysServer",
	}

	form.OnCancel = func() {
		infoLabel.Text = "Stopping sysServer"
		sysServer.Stop()
		infoLabel.Text = "sysServer Stopped"
		form.SubmitText = "Start sysServer"
		oscOutput.Enable()
		oscInput.Enable()
		oscAddr.Enable()
		form.Refresh()
	}
	form.OnSubmit = func() {
		clipPath = path.Text
		OSCOutPort = oscOutput.Text
		OSCPort = oscInput.Text
		OSCAddr = oscAddr.Text
		infoLabel.Text = "Starting sysServer"

		sysServer.Start()

		infoLabel.Text = fmt.Sprintf("sysServer Started. Open your web browser to: http://%s:%s", getIP().String(), httpPort)
		form.SubmitText = "Update sysServer"
		broadcast.Publish("/path ,s " + clipPath)
		oscOutput.Disable()
		oscInput.Disable()
		oscAddr.Disable()
		form.Refresh()
	}

	w.SetContent(container.NewGridWithRows(2, form, infoLabel))
	w.ShowAndRun()
}
