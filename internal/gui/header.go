package gui

import (
	d "github.com/lxn/walk/declarative"
)

func header() d.Composite {
	return d.Composite{
		Layout: d.HBox{MarginsZero: true},
		Children: []d.Widget{
			d.Composite{
				Layout: d.VBox{
					MarginsZero: true,
					Spacing:     6,
				},
				Children: []d.Widget{
					d.Label{
						Text:          "Safecracker Fix Utility",
						TextAlignment: d.AlignNear,
						Font:          d.Font{Family: "Segoe UI Semibold", PointSize: 14},
					},
					d.Label{
						Text:          "Performance and graphics compatibility fixes",
						TextAlignment: d.AlignNear,
					},
				},
			},
			d.HSpacer{},
		},
	}
}
