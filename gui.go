package main

import (
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"runtime"
)

func gui() {
	a := app.New()
	w := a.NewWindow("Timecode Monitor Server")

	infoLabel := widget.NewLabel("Server Stopped")

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

	httpPortField := widget.NewEntry()
	httpPortField.SetText(httpPort)
	httpPortField.Validator = validation.NewRegexp(`^[0-9]*$`, "not a valid port")

	messageField := widget.NewEntry()
	messageField.SetText(clientMessage)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Path", Widget: path, HintText: "OSC Path for clip to listen to"},
			{Text: "OSC Input Port", Widget: oscInput, HintText: "OSC Input port (usually 7000)"},
			{Text: "OSC Output Port", Widget: oscOutput, HintText: "OSC Output port (usually 7001) Note: If you have multiple services using Resolume OSC make use the correct broadcast address."},
			{Text: "OSC Host Address", Widget: oscAddr, HintText: "IP address of device that's running Resolume (make sure to open the OSC input port in your firewall)"},
			{Text: "HTTP Server Port", Widget: httpPortField, HintText: "The port to run the browser interface on"},
			{Text: "Message to client", Widget: messageField, HintText: "A message to send to all clients"},
		},
		SubmitText: "Start Server",
		CancelText: "Stop Server",
	}

	form.OnCancel = nil
	form.OnSubmit = func() {
		clipPath = path.Text
		OSCOutPort = oscOutput.Text
		OSCPort = oscInput.Text
		OSCAddr = oscAddr.Text
		httpPort = httpPortField.Text

		if messageField.Text != clientMessage {
			clientMessage = messageField.Text
			pushClientMessage()
		}

		infoLabel.SetText("Starting Server")

		if err := serverStart(); err != nil {
			dialog.ShowError(err, w)
			infoLabel.SetText("Server Errored")
			return
		}

		infoLabel.SetText(fmt.Sprintf("Server Started. Open your web browser to: http://%s:%s", getIP().String(), httpPort))
		form.SubmitText = "Update Server"
		oscOutput.Disable()
		oscInput.Disable()
		oscAddr.Disable()
		httpPortField.Disable()

		form.OnCancel = func() {
			infoLabel.SetText("Stopping Server")
			serverStop()
			infoLabel.SetText("Server Stopped")
			form.SubmitText = "Start Server"
			oscOutput.Enable()
			oscInput.Enable()
			oscAddr.Enable()
			httpPortField.Enable()

			form.OnCancel = nil

			form.Refresh()
			runtime.GC()
		}

		form.Refresh()
		runtime.GC()
	}

	w.SetContent(container.NewGridWithRows(2, form, infoLabel))
	w.ShowAndRun()
}
