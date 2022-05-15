package gui

import (
	_ "embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"html/template"
	"resolume-timecode/config"
	"resolume-timecode/services"
	"resolume-timecode/services/clients/gui"
	"resolume-timecode/services/clients/html"
	"resolume-timecode/services/clients/osc"
	"resolume-timecode/services/server"
	"resolume-timecode/util"
	"runtime"
)

func Gui(a fyne.App, logo *fyne.StaticResource) {
	w := a.NewWindow("Timecode Monitor Server")
	w.SetIcon(logo)

	infoLabel := widget.NewRichTextWithText("Server Stopped")
	infoLabel.Wrapping = fyne.TextWrapWord

	serverForm, updateServerForm, enableServerForm := genServerForm()
	configForm, updateConfigForm, enableConfigForm := genConfigForm()
	clientForm, updateClientForm := genClientForm() // TODO: fix disabled state
	statusBar, enableReset := genStatusBar()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Widget: NewValidateTabs(
				container.NewTabItem("Client Settings", clientForm),
				container.NewTabItem("Clients", configForm),
				container.NewTabItem("Server Settings", serverForm),
			),
			},
		},
		SubmitText: "Start Server",
		CancelText: "Stop Server",
	}

	form.OnCancel = nil
	form.OnSubmit = func() {
		updateClientForm()
		updateConfigForm()
		updateServerForm()

		config.StoreValues()

		infoLabel.ParseMarkdown("Starting Server")

		if err := services.Start(); err != nil {
			dialog.ShowError(err, w)
			infoLabel.ParseMarkdown("Server Errored")
			return
		}

		ip, err := util.ExternalIP()
		if err != nil {
			dialog.ShowError(err, w)
			services.Stop()
			infoLabel.ParseMarkdown("Server Errored")
			return
		}

		infoLabel.ParseMarkdown(fmt.Sprintf("Server Running. Open your web browser to [http://%s:%s](http://%[1]s:%[2]s/) (or any other address for this device) to view the timecode.\n", ip, config.GetString(config.HTTPPort)))
		form.SubmitText = "Update Settings"

		enableServerForm(false)
		enableConfigForm(false)
		enableReset(true)

		server.Reset()

		form.OnCancel = func() {
			infoLabel.ParseMarkdown("Stopping Server")

			enableReset(false)
			services.Stop()

			infoLabel.ParseMarkdown("Server Stopped")
			form.SubmitText = "Start Server"

			enableServerForm(true)
			enableConfigForm(true)

			form.OnCancel = nil

			form.Refresh()
			runtime.GC()
		}

		form.Refresh()
		runtime.GC()
	}

	w.SetContent(container.NewVSplit(form,
		container.NewBorder(infoLabel, statusBar, nil, nil)))
	w.ShowAndRun()
}

func genServerForm() (*widget.Form, func(), func(bool)) {
	oscOutput := widget.NewEntry()
	oscOutput.SetText(config.GetString(config.OSCOutPort))
	oscOutput.Validator = validation.NewRegexp(`^[0-9]+$`, "not a valid port")

	oscInput := widget.NewEntry()
	oscInput.SetText(config.GetString(config.OSCPort))
	oscInput.Validator = validation.NewRegexp(`^[0-9]+$`, "not a valid port")

	oscAddr := widget.NewEntry()
	oscAddr.SetText(config.GetString(config.OSCAddr))
	oscAddr.Validator = validation.NewRegexp(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, "not a valid IP address")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "OSC Input Port", Widget: oscInput, HintText: "OSC Input port (usually 7000)"},
			{Text: "OSC Output Port", Widget: oscOutput, HintText: "OSC Output port (usually 7001) Note: If you have multiple services using Resolume OSC make use the correct broadcast address."},
			{Text: "OSC Host Address", Widget: oscAddr, HintText: "IP address of device that's running Resolume (make sure to open the OSC input port in your firewall)"},
		},
	}

	save := func() {
		config.SetString(config.OSCOutPort, oscOutput.Text)
		config.SetString(config.OSCPort, oscInput.Text)
		config.SetString(config.OSCAddr, oscAddr.Text)
	}

	enable := func(enable bool) {
		if enable {
			oscOutput.Enable()
			oscInput.Enable()
			oscAddr.Enable()
		} else {
			oscOutput.Disable()
			oscInput.Disable()
			oscAddr.Disable()
		}
	}

	return form, save, enable
}

func genConfigForm() (*widget.Form, func(), func(bool)) {
	form := &widget.Form{
		Items: []*widget.FormItem{},
	}

	htmlForm, htmlSave, htmlEnable := html.GenHtmlClientGui()
	oscForm, oscSave, oscEnable := osc.GenOSCClientGui()

	form.Items = append(form.Items, htmlForm...)
	form.Items = append(form.Items, oscForm...)

	save := func() {
		htmlSave()
		oscSave()
	}

	enable := func(b bool) {
		htmlEnable(b)
		oscEnable(b)
	}

	enable(false)
	enable(true)

	return form, save, enable
}

func genClientForm() (*widget.Form, func()) {
	path := widget.NewSelectEntry([]string{"", "/composition/selectedclip", "/composition/layers/1/clips/1", "/composition/selectedlayer", "/composition/layers/1"})
	path.SetText(config.GetString(config.ClipPath))
	path.SetPlaceHolder("Path to clip (/composition/...)")
	path.Validator = validation.NewRegexp(`^/[^\?\,\[\]\{\}\#\s]+$`, "not a valid OSC path")

	messageField := widget.NewEntry()
	messageField.SetPlaceHolder("Message to send (optional)")

	invertField := widget.NewCheck("", nil)
	invertField.SetChecked(!config.GetBool(config.ClipInvert))

	alertTimeLabel := widget.NewLabel(fmt.Sprintf("%02d", config.GetInt(config.AlertTime)))

	alertTimeField := widget.NewSlider(0, 20)
	alertTimeField.SetValue(float64(config.GetInt(config.AlertTime)))
	alertTimeField.Step = 1
	alertTimeField.OnChanged = func(v float64) {
		alertTimeLabel.SetText(fmt.Sprintf("%02d", int(v)))
	}

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Path", Widget: path, HintText: "OSC Path for clip to listen to"},
			{Text: "Message to client", Widget: messageField, HintText: "A message to send to all clients"},
			{Text: "Use T-", Widget: invertField, HintText: "Use T- instead of T+"},
			{Text: "Alert Time", Widget: container.NewBorder(nil, nil, alertTimeLabel, nil, alertTimeField), HintText: "Time in seconds before the end of a video to alert the client"},
		},
	}

	save := func() {
		config.SetString(config.ClipPath, path.Text)
		config.SetBool(config.ClipInvert, !invertField.Checked)
		config.SetString(config.ClientMessage, template.HTMLEscapeString(messageField.Text))
		config.SetInt(config.AlertTime, int(alertTimeField.Value))
	}

	return form, save
}

func genStatusBar() (*fyne.Container, func(bool)) {
	timeLeftLabel := widget.NewLabelWithData(gui.TimeLeftBinding)
	clipLengthLabel := widget.NewLabelWithData(gui.ClipLengthBinding)
	clipNameLabel := widget.NewLabelWithData(gui.ClipNameBinding)
	clipNameLabel.Wrapping = fyne.TextTruncate

	resetButton := widget.NewButton("Reset Timecode", server.Reset)
	resetButton.Disable()

	enable := func(enable bool) {
		if enable {
			resetButton.Enable()
		} else {
			resetButton.Disable()
		}
	}

	return container.NewGridWithColumns(4, timeLeftLabel, clipLengthLabel, clipNameLabel, resetButton), enable
}
