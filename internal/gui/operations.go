package gui

func (w *Window) beginOperation(status string) {
	w.busy = true

	w.directoryEdit.SetEnabled(false)
	w.browseButton.SetEnabled(false)
	w.performanceFix.button.SetEnabled(false)
	w.graphicsFix.button.SetEnabled(false)
	w.statusLabel.SetText(status)
}

func (w *Window) endOperation() {
	w.busy = false

	w.directoryEdit.SetEnabled(true)
	w.browseButton.SetEnabled(true)
}
