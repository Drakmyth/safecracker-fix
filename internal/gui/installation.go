package gui

import (
	"strings"
	"time"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

const DIRECTORY_VALIDATION_DELAY = 500 * time.Millisecond

func (w *Window) installationSection() d.GroupBox {
	return d.GroupBox{
		Title: "Game Installation",
		Layout: d.VBox{
			Margins: d.Margins{Left: 12, Top: 12, Right: 12, Bottom: 12},
			Spacing: 10,
		},
		Children: []d.Widget{
			w.directoryRow(),
			mutedLabel("No installation selected", &w.targetLabel),
		},
	}
}

func (w *Window) directoryRow() d.Composite {
	return d.Composite{
		Layout: d.HBox{MarginsZero: true},
		Children: []d.Widget{
			d.LineEdit{
				AssignTo:          &w.directoryEdit,
				OnTextChanged:     w.directoryChanged,
				OnEditingFinished: w.directoryEditingFinished,
			},
			d.PushButton{
				AssignTo:  &w.browseButton,
				Text:      "Browse...",
				MinSize:   d.Size{Width: 80, Height: 32},
				MaxSize:   d.Size{Width: 80},
				OnClicked: w.browseDirectory,
			},
		},
	}
}

func (w *Window) browseDirectory() {
	directory, accepted, err := browseFolder(w.mainWindow, "Select Safecracker Installation Directory", w.directoryPath())

	if err != nil {
		walk.MsgBox(w.mainWindow, "Directory Selection Error", err.Error(), walk.MsgBoxIconError)
		return
	}

	if !accepted {
		return
	}

	w.directoryEdit.SetText(directory)
	w.directoryEditingFinished()
}

func (w *Window) directoryChanged() {
	w.cancelDirectoryValidation()

	w.installation = nil

	w.performanceStatus.SetText("Unknown")
	w.performanceIcon.SetImage(w.statusIcons[STATUS_ICON_UNKNOWN])
	w.performanceButton.SetText("Apply")
	w.performanceButton.SetEnabled(false)

	w.graphicsStatus.SetText("Unknown")
	w.graphicsIcon.SetImage(w.statusIcons[STATUS_ICON_UNKNOWN])
	w.graphicsButton.SetText("Apply")
	w.graphicsButton.SetEnabled(false)

	directory := w.directoryPath()

	if directory == "" {
		w.targetLabel.SetText("No installation selected")
		return
	}

	w.targetLabel.SetText("Waiting to check installation...")
	w.scheduleDirectoryValidation(directory)
}

func (w *Window) validateDirectory(directory string) {
	w.targetLabel.SetText("Checking installation...")
}

func (w *Window) cancelDirectoryValidation() {
	w.directoryRevision++

	if w.directoryTimer != nil {
		w.directoryTimer.Stop()
		w.directoryTimer = nil
	}
}

func (w *Window) scheduleDirectoryValidation(directory string) {
	revision := w.directoryRevision

	w.directoryTimer = time.AfterFunc(DIRECTORY_VALIDATION_DELAY, func() {
		w.mainWindow.Synchronize(func() {
			if revision != w.directoryRevision {
				return
			}

			w.directoryTimer = nil
			w.validateDirectory(directory)
		})
	})
}

func (w *Window) directoryEditingFinished() {
	if w.directoryTimer == nil {
		return
	}

	w.cancelDirectoryValidation()
	w.validateDirectory(w.directoryPath())
}

func (w *Window) directoryPath() string {
	return strings.TrimSpace(w.directoryEdit.Text())
}
