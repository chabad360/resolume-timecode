package main

import (
	_ "embed"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/chabad360/go-osc/osc"
	"html/template"
	"runtime"
)

var (
	//go:embed images/logo.png
	logo              []byte
	logoResource      = fyne.NewStaticResource("logo", logo)
	clipLengthBinding = binding.NewString()
	timeLeftBinding   = binding.NewString()
	clipNameBinding   = binding.NewString()
)

var _ fyne.Widget = (*ValidateTabs)(nil)
var _ fyne.Validatable = (*ValidateTabs)(nil)

type ValidateTabs struct {
	*container.AppTabs
	f   func(error)
	err error
}

func (v *ValidateTabs) Validate() error {
	for _, item := range v.Items {
		if w, ok := item.Content.(fyne.Validatable); ok {
			if err := w.Validate(); err != nil {
				//v.setValidationError(err)
				return err
			}
		}
	}
	return nil
}

func (v *ValidateTabs) SetValidationError(err error) {
	if err == nil && v.err == nil {
		return
	}

	if (err == nil && v.err != nil) || (v.err == nil && err != nil) ||
		err.Error() != v.err.Error() {
		if err == nil {
			for _, item := range v.Items {
				if w, ok := item.Content.(fyne.Validatable); ok {
					w.SetOnValidationChanged(func(_ error) {}) // prevent recursion
					e := w.Validate()
					w.SetOnValidationChanged(func(err error) {
						if err != nil {
							item.Icon = theme.ErrorIcon()
						} else {
							item.Icon = theme.ConfirmIcon()
						}
						v.SetValidationError(err)
					})
					if e != nil {
						err = e
						break
					}
				}
			}
		}
		v.err = err

		if v.f != nil {
			v.f(err)
		}
	}
}

func (v *ValidateTabs) SetOnValidationChanged(f func(error)) {
	if f != nil {
		v.f = f
	}
}

func NewValidateTabs(tabs ...*container.TabItem) *ValidateTabs {
	v := &ValidateTabs{AppTabs: container.NewAppTabs(tabs...), err: errors.New("validation failed")}
	for _, item := range v.Items {
		if w, ok := item.Content.(fyne.Validatable); ok {
			w.SetOnValidationChanged(func(err error) {
				if err != nil {
					item.Icon = theme.ErrorIcon()
				} else {
					item.Icon = theme.ConfirmIcon()
				}
				v.SetValidationError(err)
			})
		}
	}
	return v
}
func gui() {
	w := a.NewWindow("Timecode Monitor Server")
	w.SetIcon(logoResource)

	infoLabel := widget.NewRichTextWithText("Server Stopped")
	infoLabel.Wrapping = fyne.TextWrapWord

	timeLeftLabel := widget.NewLabelWithData(timeLeftBinding)
	clipLengthLabel := widget.NewLabelWithData(clipLengthBinding)
	clipNameLabel := widget.NewLabelWithData(clipNameBinding)
	clipNameLabel.Wrapping = fyne.TextTruncate

	resetButton := widget.NewButton("Reset Timecode", reset)
	resetButton.Hide()

	path := widget.NewSelectEntry([]string{"", "/composition/selectedclip", "/composition/layers/1/clips/1", "/composition/selectedlayer", "/composition/layers/1"})
	path.SetText(clipPath)
	path.SetPlaceHolder("Path to clip (/composition/...)")
	path.Validator = validation.NewRegexp(`^[^\?\,\[\]\{\}\#\s]+$`, "not a valid OSC path")

	oscOutput := widget.NewEntry()
	oscOutput.SetText(OSCOutPort)
	oscOutput.Validator = validation.NewRegexp(`^[0-9]+$`, "not a valid port")

	oscInput := widget.NewEntry()
	oscInput.SetText(OSCPort)
	oscInput.Validator = validation.NewRegexp(`^[0-9]+$`, "not a valid port")

	oscAddr := widget.NewEntry()
	oscAddr.SetText(OSCAddr)
	oscAddr.Validator = validation.NewRegexp(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, "not a valid IP address")

	httpPortField := widget.NewEntry()
	httpPortField.SetText(httpPort)
	httpPortField.Validator = validation.NewRegexp(`^[0-9]+$`, "not a valid port")

	messageField := widget.NewEntry()
	messageField.SetText(clientMessage)

	invertField := widget.NewCheck("", nil)
	invertField.SetChecked(!clipInvert)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Widget: NewValidateTabs(container.NewTabItemWithIcon("Client Settings", theme.ConfirmIcon(), &widget.Form{
				Items: []*widget.FormItem{
					{Text: "Path", Widget: path, HintText: "OSC Path for clip to listen to"},
					{Text: "Message to client", Widget: messageField, HintText: "A message to send to all clients"},
					{Text: "Use T-", Widget: invertField, HintText: "Use T- instead of T+"},
				},
			}),
				container.NewTabItemWithIcon("Server Settings", theme.ConfirmIcon(), &widget.Form{
					Items: []*widget.FormItem{
						{Text: "OSC Input Port", Widget: oscInput, HintText: "OSC Input port (usually 7000)"},
						{Text: "OSC Output Port", Widget: oscOutput, HintText: "OSC Output port (usually 7001) Note: If you have multiple services using Resolume OSC make use the correct broadcast address."},
						{Text: "OSC Host Address", Widget: oscAddr, HintText: "IP address of device that's running Resolume (make sure to open the OSC input port in your firewall)"},
						{Text: "HTTP Server Port", Widget: httpPortField, HintText: "The port to run the browser interface on"},
					},
				}),
			),
			},
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
		broadcast.Publish(osc.NewMessage("/message", clientMessage))

		clipInvert = !invertField.Checked
		a.Preferences().SetBool("clipInvert", clipInvert)
		broadcast.Publish(osc.NewMessage("/tminus", !clipInvert))

		infoLabel.ParseMarkdown("Starting Server")

		if err := serverStart(); err != nil {
			dialog.ShowError(err, w)
			infoLabel.ParseMarkdown("Server Errored")
			return
		}

		reset()

		ip, err := externalIP()
		if err != nil {
			dialog.ShowError(err, w)
			infoLabel.ParseMarkdown("Server Errored")
			return
		}

		infoLabel.ParseMarkdown(fmt.Sprintf("Server Running. Open your web browser to [http://%s:%s](http://%[1]s:%[2]s/) (or any other address for this device) to view the timecode.\n", ip, httpPort))
		form.SubmitText = "Update Settings"
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
