package osc

import (
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
	"resolume-timecode/config"
)

func GenOSCClientGui() ([]*widget.FormItem, func(), func(bool)) {
	oscClientAddr := widget.NewEntry()
	oscClientAddr.SetPlaceHolder("192.168.1.1")
	oscClientAddr.SetText(config.GetString(config.OSCClientAddr))
	oscClientAddr.Validator = validation.NewRegexp(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`, "not a valid IP address")

	oscPortField := widget.NewEntry()
	oscPortField.SetText(config.GetString(config.OSCClientPort))
	oscPortField.Validator = validation.NewRegexp(`^[0-9]+$`, "not a valid port")

	oscEnable := widget.NewCheck("Enable", func(b bool) {
		if b {
			oscClientAddr.Enable()
			oscClientAddr.Validate()
			oscPortField.Enable()
			oscPortField.Validate()
		} else {
			oscPortField.SetValidationError(nil)
			oscPortField.Disable()
			oscClientAddr.SetValidationError(nil)
			oscClientAddr.Disable()
		}
	})
	oscEnable.SetChecked(config.GetBool(config.EnableOSCClient))

	items := []*widget.FormItem{
		{Text: "Enable OSC Client", Widget: oscEnable, HintText: "Use an OSC server (i.e. the included TouchOSC panel)"},
		{Text: "OSC Client Address", Widget: oscClientAddr, HintText: "The IP address of the OSC server"},
		{Text: "OSC Server Port", Widget: oscPortField, HintText: "The port of the OSC server"},
	}

	save := func() {
		config.SetBool(config.EnableOSCClient, oscEnable.Checked)
		config.SetString(config.OSCClientAddr, oscClientAddr.Text)
		config.SetString(config.OSCClientPort, oscPortField.Text)
	}

	enable := func(enable bool) {
		if enable {
			oscEnable.Enable()
			if oscEnable.Checked {
				oscClientAddr.Enable()
				oscPortField.Enable()
			}
		} else {
			oscEnable.Disable()
			oscClientAddr.Disable()
			oscPortField.Disable()
		}
	}

	return items, save, enable
}
