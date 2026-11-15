package gui

import "github.com/lxn/walk"

func (w *Window) beginOperation(status string) {
	w.busy = true

	w.directoryEdit.SetEnabled(false)
	w.browseButton.SetEnabled(false)
	w.performanceFix.button.SetEnabled(false)
	w.graphicsFix.button.SetEnabled(false)
	w.statusLabel.SetText(status)
}

func (w *Window) runOperation(status string, errorTitle string, operation func() error) {
	w.beginOperation(status)

	go func() {
		err := operation()

		w.mainWindow.Synchronize(func() {
			w.endOperation()

			if err != nil {
				walk.MsgBox(w.mainWindow, errorTitle, err.Error(), walk.MsgBoxIconError)
			}

			w.requestImmediateDirectoryValidation()
		})
	}()
}

func (w *Window) endOperation() {
	w.busy = false

	w.directoryEdit.SetEnabled(true)
	w.browseButton.SetEnabled(true)
}
