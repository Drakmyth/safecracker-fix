package gui

import (
	"github.com/Drakmyth/safecracker-fix/internal/installer"
	"github.com/Drakmyth/safecracker-fix/internal/patcher"
	"github.com/Drakmyth/safecracker-fix/internal/safecracker"
	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

type fixControls struct {
	status *walk.Label
	icon   *walk.ImageView
	button *walk.PushButton
}

func (f *fixControls) configure(status string, icon *walk.Bitmap, buttonText string, enabled bool) {
	textColor := win.COLOR_BTNTEXT
	if !enabled {
		textColor = win.COLOR_GRAYTEXT
	}

	f.status.SetText(status)
	f.status.SetTextColor(walk.Color(win.GetSysColor(textColor)))
	f.icon.SetImage(icon)
	f.button.SetText(buttonText)
	f.button.SetEnabled(enabled)
}

func (w *Window) fixesSection() d.GroupBox {
	return d.GroupBox{
		Title: "Available Fixes",
		Layout: d.VBox{
			Margins: d.Margins{Left: 12, Top: 12, Right: 12, Bottom: 12},
			Spacing: 14,
		},
		Children: []d.Widget{
			fixRow(
				"Graphics Compatibility Fix",
				"Resolves rendering issues using dgVoodoo2.",
				&w.graphicsFix,
				w.statusIcons[STATUS_ICON_UNKNOWN],
				w.toggleGraphicsFix,
			),
			d.VSeparator{},
			fixRow(
				"High Uptime Performance Fix",
				"Fixes extremely low frame rates.",
				&w.performanceFix,
				w.statusIcons[STATUS_ICON_UNKNOWN],
				w.togglePerformanceFix,
			),
		},
	}
}

func fixRow(
	name string,
	description string,
	controls *fixControls,
	initialIcon *walk.Bitmap,
	onClicked walk.EventHandler,
) d.Composite {
	return d.Composite{
		Layout: d.HBox{
			MarginsZero: true,
			Spacing:     12,
		},
		Children: []d.Widget{
			d.ImageView{
				AssignTo: &controls.icon,
				Image:    initialIcon,
				Mode:     d.ImageViewModeCenter,
				MinSize:  d.Size{Width: 32, Height: 32},
				MaxSize:  d.Size{Width: 32, Height: 32},
			},
			fixDetails(name, description, &controls.status),
			d.HSpacer{},
			d.PushButton{
				AssignTo:  &controls.button,
				Text:      "Apply",
				Enabled:   false,
				OnClicked: onClicked,
				MinSize:   d.Size{Width: 90, Height: 34},
				MaxSize:   d.Size{Width: 90},
			},
		},
	}
}

func fixDetails(name string, description string, status **walk.Label) d.Composite {
	return d.Composite{
		Layout: d.VBox{
			MarginsZero: true,
			Spacing:     4,
		},
		Children: []d.Widget{
			d.Label{
				Text: name,
				Font: d.Font{Family: "Segoe UI Semibold", PointSize: 10},
			},
			d.Label{Text: description},
			mutedLabel("Unknown", status),
		},
	}
}

func (w *Window) refreshPerformanceFix(inspection *patcher.Inspection) {
	state := inspection.Patches[safecracker.PATCH_FPU_PRESERVE]

	var status string
	var icon StatusIcon
	var buttonText string
	var enabled bool

	switch state {
	case patcher.PATCH_STATE_UNAPPLIED:
		status = "Not applied"
		icon = STATUS_ICON_NOT_APPLIED
		buttonText = "Apply"
		enabled = true
	case patcher.PATCH_STATE_APPLIED:
		status = "Applied"
		icon = STATUS_ICON_APPLIED
		buttonText = "Remove"
		enabled = true
	case patcher.PATCH_STATE_PARTIAL:
		status = "Incomplete"
		icon = STATUS_ICON_INCOMPLETE
		buttonText = "Apply"
		enabled = true
	default:
		status = "Unknown"
		icon = STATUS_ICON_UNKNOWN
		buttonText = "Apply"
		enabled = false
	}

	w.performanceFix.configure(status, w.statusIcons[icon], buttonText, enabled)
}

func (w *Window) refreshGraphicsFix(inspection *installer.GraphicsShimInspection) {
	var status string
	var icon StatusIcon
	var buttonText string
	var enabled bool

	switch inspection.State {
	case installer.GRAPHICS_SHIM_STATE_NOT_INSTALLED:
		status = "Not applied"
		icon = STATUS_ICON_NOT_APPLIED
		buttonText = "Apply"
		enabled = true
	case installer.GRAPHICS_SHIM_STATE_MANAGED:
		status = "Applied"
		icon = STATUS_ICON_APPLIED
		buttonText = "Remove"
		enabled = true
	case installer.GRAPHICS_SHIM_STATE_EXTERNAL:
		status = "Applied (external)"
		icon = STATUS_ICON_EXTERNAL
		buttonText = "Remove"
		enabled = false
	case installer.GRAPHICS_SHIM_STATE_INCOMPLETE:
		status = "Incomplete"
		icon = STATUS_ICON_INCOMPLETE
		buttonText = "Apply"
		enabled = true
	case installer.GRAPHICS_SHIM_STATE_CONFLICT:
		status = "Conflict"
		icon = STATUS_ICON_CONFLICT
		buttonText = "Apply"
		enabled = false
	default:
		status = "Unknown"
		icon = STATUS_ICON_UNKNOWN
		buttonText = "Apply"
		enabled = false
	}

	w.graphicsFix.configure(status, w.statusIcons[icon], buttonText, enabled)
}

func (w *Window) togglePerformanceFix() {
	if w.busy || w.installation == nil {
		return
	}

	installation := w.installation
	state := installation.Inspection.Patches[safecracker.PATCH_FPU_PRESERVE]

	switch state {
	case patcher.PATCH_STATE_APPLIED:
		w.runOperation("Removing performance fix...", "Performance Fix Error", func() error {
			_, err := installation.Reverse(safecracker.PATCH_FPU_PRESERVE)
			return err
		})
	case patcher.PATCH_STATE_UNAPPLIED, patcher.PATCH_STATE_PARTIAL:
		w.runOperation("Applying performance fix...", "Performance Fix Error", func() error {
			_, err := installation.Apply(safecracker.PATCH_FPU_PRESERVE)
			return err
		})
	}
}

func (w *Window) toggleGraphicsFix() {
	if w.busy || w.installation == nil || w.graphicsInspection == nil {
		return
	}

	installation := w.installation

	switch w.graphicsInspection.State {
	case installer.GRAPHICS_SHIM_STATE_MANAGED:
		w.runOperation("Removing graphics fix...", "Graphics Fix Error", func() error {
			_, err := installation.RemoveGraphicsShim()
			return err
		})
	case installer.GRAPHICS_SHIM_STATE_NOT_INSTALLED, installer.GRAPHICS_SHIM_STATE_INCOMPLETE:
		w.runOperation("Applying graphics fix...", "Graphics Fix Error", func() error {
			_, err := installation.InstallGraphicsShim()
			return err
		})
	}
}
