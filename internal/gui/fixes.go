package gui

import (
	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

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
				&w.graphicsIcon,
				w.statusIcons[STATUS_ICON_UNKNOWN],
				&w.graphicsStatus,
				&w.graphicsButton,
			),
			d.VSeparator{},
			fixRow(
				"High Uptime Performance Fix",
				"Fixes extremely low frame rates.",
				&w.performanceIcon,
				w.statusIcons[STATUS_ICON_UNKNOWN],
				&w.performanceStatus,
				&w.performanceButton,
			),
		},
	}
}

func fixRow(
	name string,
	description string,
	icon **walk.ImageView,
	initialIcon *walk.Bitmap,
	status **walk.Label,
	button **walk.PushButton,
) d.Composite {
	return d.Composite{
		Layout: d.HBox{
			MarginsZero: true,
			Spacing:     12,
		},
		Children: []d.Widget{
			d.ImageView{
				AssignTo: icon,
				Image:    initialIcon,
				Mode:     d.ImageViewModeCenter,
				MinSize:  d.Size{Width: 32, Height: 32},
				MaxSize:  d.Size{Width: 32, Height: 32},
			},
			fixDetails(name, description, status),
			d.HSpacer{},
			d.PushButton{
				AssignTo: button,
				Text:     "Apply",
				Enabled:  false,
				MinSize:  d.Size{Width: 90, Height: 34},
				MaxSize:  d.Size{Width: 90},
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
