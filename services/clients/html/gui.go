package html

import (
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
	"resolume-timecode/config"
)

func GenHtmlClientGui() ([]*widget.FormItem, func(), func(bool)) {
	httpPortField := widget.NewEntry()
	httpPortField.SetText(config.GetString(config.HTTPPort))
	httpPortField.Validator = validation.NewRegexp(`^[0-9]+$`, "not a valid port")

	httpEnable := widget.NewCheck("Enable", func(b bool) {
		if b {
			httpPortField.Enable()
			httpPortField.Validate()
		} else {
			httpPortField.SetValidationError(nil)
			httpPortField.Disable()
		}
	})
	httpEnable.SetChecked(config.GetBool(config.EnableHttpClient))

	items := []*widget.FormItem{
		{Text: "Enable HTTP Client", Widget: httpEnable, HintText: "Use the HTTP browser client"},
		{Text: "HTTP Server Port", Widget: httpPortField, HintText: "The port to run the browser interface on"},
	}

	save := func() {
		config.SetBool(config.EnableHttpClient, httpEnable.Checked)
		config.SetString(config.HTTPPort, httpPortField.Text)
	}

	enable := func(enable bool) {
		if enable {
			httpEnable.Enable()
			if httpEnable.Checked {
				httpPortField.Enable()
			}
		} else {
			httpEnable.Disable()
			httpPortField.Disable()
		}
	}

	return items, save, enable
}
