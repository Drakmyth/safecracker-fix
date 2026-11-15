package gui

import (
	"time"

	"github.com/Drakmyth/safecracker-fix/internal/installer"
	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

type Window struct {
	mainWindow *walk.MainWindow
	version    string
	closing    bool
	busy       bool

	directoryEdit *walk.LineEdit
	browseButton  *walk.PushButton
	targetLabel   *walk.Label

	performanceFix fixControls
	graphicsFix    fixControls

	statusIcons map[StatusIcon]*walk.Bitmap
	statusLabel *walk.Label

	installation       *installer.Installation
	graphicsInspection *installer.GraphicsShimInspection
	directoryTimer     *time.Timer
	directoryRevision  uint64
}

func Run(version string) error {
	w := &Window{version: version}

	icons, err := loadStatusIcons()
	if err != nil {
		return err
	}
	defer disposeStatusIcons(icons)
	w.statusIcons = icons

	window := d.MainWindow{
		AssignTo: &w.mainWindow,
		Title:    "Safecracker Fix Utility",
		Size:     d.Size{Width: 640, Height: 480},
		MinSize:  d.Size{Width: 640, Height: 480},
		MaxSize:  d.Size{Width: 640, Height: 480},
		Font:     d.Font{Family: "Segoe UI", PointSize: 10},
		Layout: d.VBox{
			Margins: d.Margins{Left: 16, Top: 16, Right: 16, Bottom: 12},
			Spacing: 12,
		},
		Children: []d.Widget{
			header(),
			w.installationSection(),
			w.fixesSection(),
			w.statusBar(),
		},
	}

	if err := window.Create(); err != nil {
		return err
	}

	w.mainWindow.Closing().Attach(
		func(canceled *bool, reason walk.CloseReason) {
			if *canceled {
				return
			}

			if w.busy {
				*canceled = true
				return
			}

			w.closing = true
			w.cancelDirectoryValidation()
		},
	)

	hwnd := w.mainWindow.Handle()

	style := win.GetWindowLong(hwnd, win.GWL_STYLE)
	style &^= win.WS_THICKFRAME | win.WS_MAXIMIZEBOX
	win.SetWindowLong(hwnd, win.GWL_STYLE, style)
	// Refresh the window frame after changing its styles, without moving or resizing the window.
	win.SetWindowPos(hwnd, 0, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_NOZORDER|win.SWP_FRAMECHANGED)

	w.browseButton.SetFocus()
	w.mainWindow.Run()

	return nil
}

func (w *Window) statusBar() d.Composite {
	return d.Composite{
		Layout: d.HBox{MarginsZero: true},
		Children: []d.Widget{
			mutedLabel("Ready", &w.statusLabel),
			d.HSpacer{},
			mutedLabel("v"+w.version, nil),
		},
	}
}
