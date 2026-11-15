package gui

import (
	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func mutedLabel(text string, assignTo **walk.Label) d.Label {
	return d.Label{
		AssignTo:  assignTo,
		Text:      text,
		TextColor: walk.Color(win.GetSysColor(win.COLOR_GRAYTEXT)),
		Background: d.SystemColorBrush{
			Color: walk.SysColorBtnFace,
		},
	}
}
