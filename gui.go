package main

import (
	_ "embed"
	"fmt"
	"github.com/chabad360/go-osc/osc"
	"html/template"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var (
	//go:embed images/logo.png
	logo              []byte
	logoResource      = fyne.NewStaticResource("logo", logo)
	clipLengthBinding = binding.NewString()
	timeLeftBinding   = binding.NewString()
	clipNameBinding   = binding.NewString()
)

func gui() {
	w := a.NewWindow("Timecode Monitor Server")
	w.SetIcon(logoResource)

	infoLabel := widget.NewRichTextWithText("Server Stopped")
	infoLabel.Wrapping = fyne.TextWrapOff

	timeLeftLabel := widget.NewLabelWithData(timeLeftBinding)
	clipLengthLabel := widget.NewLabelWithData(clipLengthBinding)
	clipNameLabel := widget.NewLabelWithData(clipNameBinding)
	clipNameLabel.Wrapping = fyne.TextTruncate

	resetButton := widget.NewButton("Reset Timecode", lightReset)
	resetButton.Hide()

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

	invertField := widget.NewCheck("", func(b bool) {
		clipInvert = !b
		a.Preferences().SetBool("clipInvert", clipInvert)
		broadcast.Publish(osc.NewMessage("/tminus", !clipInvert))
	})
	invertField.SetChecked(!clipInvert)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Path", Widget: path, HintText: "OSC Path for clip to listen to"},
			{Text: "OSC Input Port", Widget: oscInput, HintText: "OSC Input port (usually 7000)"},
			{Text: "OSC Output Port", Widget: oscOutput, HintText: "OSC Output port (usually 7001) Note: If you have multiple services using Resolume OSC make use the correct broadcast address."},
			{Text: "OSC Host Address", Widget: oscAddr, HintText: "IP address of device that's running Resolume (make sure to open the OSC input port in your firewall)"},
			{Text: "HTTP Server Port", Widget: httpPortField, HintText: "The port to run the browser interface on"},
			{Text: "Message to client", Widget: messageField, HintText: "A message to send to all clients"},
			{Text: "Use T-", Widget: invertField, HintText: "Use T- instead of T+"},
		},
		SubmitText: "Start Server",
		CancelText: "Stop Server",
	}

	form.OnCancel = nil
	form.OnSubmit = func() {
		clipPath = path.Text
		a.Preferences().SetString("clipPath", clipPath)
		OSCOutPort = oscOutput.Text
		a.Preferences().SetString("OSCOutPort", OSCOutPort)
		OSCPort = oscInput.Text
		a.Preferences().SetString("OSCPort", OSCPort)
		OSCAddr = oscAddr.Text
		a.Preferences().SetString("OSCAddr", OSCAddr)
		httpPort = httpPortField.Text
		a.Preferences().SetString("httpPort", httpPort)

		clientMessage = template.HTMLEscapeString(messageField.Text)
		pushClientMessage()

		infoLabel.ParseMarkdown("Starting Server")

		if err := serverStart(); err != nil {
			dialog.ShowError(err, w)
			infoLabel.ParseMarkdown("Server Errored")
			return
		}

		reset()

		infoLabel.ParseMarkdown(fmt.Sprintf("Server Running. Open your web browser to [http://%s:%s](http://%[1]s:%[2]s/) to view the timecode.\n", getIP().String(), httpPort))
		form.SubmitText = "Update Server"
		oscOutput.Disable()
		oscInput.Disable()
		oscAddr.Disable()
		httpPortField.Disable()
		resetButton.Show()

		form.OnCancel = func() {
			infoLabel.ParseMarkdown("Stopping Server")
			resetButton.Hide()
			serverStop()
			clipNameBinding.Set("Clip Name: None")
			infoLabel.ParseMarkdown("Server Stopped")
			form.SubmitText = "Start Server"
			oscOutput.Enable()
			oscInput.Enable()
			oscAddr.Enable()
			httpPortField.Enable()

			form.OnCancel = nil
			reset()

			form.Refresh()
			runtime.GC()
		}

		form.Refresh()
		runtime.GC()
	}

	w.SetContent(container.NewVSplit(form,
		container.NewBorder(infoLabel,
			container.NewGridWithColumns(4, timeLeftLabel, clipLengthLabel, clipNameLabel, resetButton), nil, nil)))
	w.ShowAndRun()
}
