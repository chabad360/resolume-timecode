package gui

import (
	_ "embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"html/template"
	"resolume-timecode/config"
	"resolume-timecode/services"
	"resolume-timecode/services/clients/gui"
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
	clientForm, updateClientForm := genClientForm()
	statusBar, enableReset := genStatusBar()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Widget: NewValidateTabs(
				container.NewTabItemWithIcon("Client Settings", theme.ConfirmIcon(), clientForm),
				container.NewTabItemWithIcon("Server Settings", theme.ConfirmIcon(), serverForm),
			),
			},
		},
		SubmitText: "Start Server",
		CancelText: "Stop Server",
	}

	form.OnCancel = nil
	form.OnSubmit = func() {
		updateClientForm()
		updateServerForm()

		infoLabel.ParseMarkdown("Starting Server")

		if err := services.Start(); err != nil {
			dialog.ShowError(err, w)
			infoLabel.ParseMarkdown("Server Errored")
			return
		}

		ip, err := util.ExternalIP()
		if err != nil {
			dialog.ShowError(err, w)
			infoLabel.ParseMarkdown("Server Errored")
			return
		}

		infoLabel.ParseMarkdown(fmt.Sprintf("Server Running. Open your web browser to [http://%s:%s](http://%[1]s:%[2]s/) (or any other address for this device) to view the timecode.\n", ip, config.GetString(config.HTTPPort)))
		form.SubmitText = "Update Settings"

		enableServerForm(false)
		enableReset(true)

		form.OnCancel = func() {
			infoLabel.ParseMarkdown("Stopping Server")

			enableReset(false)
			services.Stop()

			infoLabel.ParseMarkdown("Server Stopped")
			form.SubmitText = "Start Server"

			enableServerForm(true)

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

	httpPortField := widget.NewEntry()
	httpPortField.SetText(config.GetString(config.HTTPPort))
	httpPortField.Validator = validation.NewRegexp(`^[0-9]+$`, "not a valid port")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "OSC Input Port", Widget: oscInput, HintText: "OSC Input port (usually 7000)"},
			{Text: "OSC Output Port", Widget: oscOutput, HintText: "OSC Output port (usually 7001) Note: If you have multiple services using Resolume OSC make use the correct broadcast address."},
			{Text: "OSC Host Address", Widget: oscAddr, HintText: "IP address of device that's running Resolume (make sure to open the OSC input port in your firewall)"},
			{Text: "HTTP Server Port", Widget: httpPortField, HintText: "The port to run the browser interface on"},
		},
	}

	save := func() {
		config.SetString(config.OSCOutPort, oscOutput.Text)
		config.SetString(config.OSCPort, oscInput.Text)
		config.SetString(config.OSCAddr, oscAddr.Text)
		config.SetString(config.HTTPPort, httpPortField.Text)
	}

	enable := func(enable bool) {
		if enable {
			oscOutput.Enable()
			oscInput.Enable()
			oscAddr.Enable()
			httpPortField.Enable()
		} else {
			oscOutput.Disable()
			oscInput.Disable()
			oscAddr.Disable()
			httpPortField.Disable()
		}
	}

	return form, save, enable
}

func genClientForm() (*widget.Form, func()) {
	path := widget.NewSelectEntry([]string{"", "/composition/selectedclip", "/composition/layers/1/clips/1", "/composition/selectedlayer", "/composition/layers/1"})
	path.SetText(config.GetString(config.ClipPath))
	path.SetPlaceHolder("Path to clip (/composition/...)")
	path.Validator = validation.NewRegexp(`^[^\?\,\[\]\{\}\#\s]+$`, "not a valid OSC path")

	messageField := widget.NewEntry()
	messageField.SetPlaceHolder("Message to send (optional)")

	invertField := widget.NewCheck("", nil)
	invertField.SetChecked(!config.GetBool(config.ClipInvert))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Path", Widget: path, HintText: "OSC Path for clip to listen to"},
			{Text: "Message to client", Widget: messageField, HintText: "A message to send to all clients"},
			{Text: "Use T-", Widget: invertField, HintText: "Use T- instead of T+"},
		},
	}

	save := func() {
		config.SetString(config.ClipPath, path.Text)
		config.SetBool(config.ClipInvert, !invertField.Checked)
		config.SetString(config.ClientMessage, template.HTMLEscapeString(messageField.Text))
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
