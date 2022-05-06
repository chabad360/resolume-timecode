package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Widget = (*ValidateTabs)(nil)
var _ fyne.Validatable = (*ValidateTabs)(nil)

type ValidationStatus struct {
	invalid bool
	err     error
}

type ValidateTabs struct {
	container.AppTabs
	f       func(error)
	err     error
	invalid []ValidationStatus
}

func (v *ValidateTabs) Refresh() {
	for k, item := range v.invalid {
		v.AppTabs.Items[k].Icon = theme.ConfirmIcon()
		if item.invalid {
			v.AppTabs.Items[k].Icon = theme.ErrorIcon()
		}
	}

	v.AppTabs.Refresh()
}

func (v *ValidateTabs) Validate() error {
	for _, item := range v.Items {
		if w, ok := item.Content.(fyne.Validatable); ok {
			if err := w.Validate(); err != nil {
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

	v.Refresh()

	if (err == nil && v.err != nil) || (v.err == nil && err != nil) {
		if err == nil {
			for _, item := range v.invalid {
				if item.invalid {
					err = item.err
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
	v := &ValidateTabs{}
	v.ExtendBaseWidget(v)

	v.Items = tabs
	v.invalid = make([]ValidationStatus, len(tabs))
	return v
}

func (v *ValidateTabs) CreateRenderer() fyne.WidgetRenderer {
	for k, item := range v.Items {
		if w, ok := item.Content.(fyne.Validatable); ok {
			w.SetOnValidationChanged(updateValidation(v, k))
		}
	}

	return v.AppTabs.CreateRenderer()
}

func updateValidation(v *ValidateTabs, i int) func(error) {
	return func(err error) {
		i := i
		if err != nil {
			v.invalid[i].invalid = true
			v.invalid[i].err = err
		} else {
			v.invalid[i].invalid = false
			v.invalid[i].err = nil
		}
		v.SetValidationError(err)
	}
}
