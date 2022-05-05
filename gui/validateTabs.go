package gui

import (
	"errors"
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
	*container.AppTabs
	f       func(error)
	err     error
	invalid []ValidationStatus
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
	v := &ValidateTabs{AppTabs: container.NewAppTabs(tabs...), err: errors.New("initial error"), invalid: make([]ValidationStatus, len(tabs))}
	for k, item := range v.Items {
		if w, ok := item.Content.(fyne.Validatable); ok {
			w.SetOnValidationChanged(func(err error) {
				if err != nil {
					item.Icon = theme.ErrorIcon()
					v.invalid[k].invalid = true
					v.invalid[k].err = err
				} else {
					item.Icon = theme.ConfirmIcon()
					v.invalid[k].invalid = false
					v.invalid[k].err = nil
				}
				v.SetValidationError(err)
			})
		}
	}
	return v
}
