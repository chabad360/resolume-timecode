package gui

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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
	v := &ValidateTabs{AppTabs: container.NewAppTabs(tabs...), err: errors.New("initial error")}
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
