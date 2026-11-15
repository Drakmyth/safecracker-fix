package gui

import (
	"bytes"
	"embed"
	"fmt"
	"image/png"
	"path"

	"github.com/lxn/walk"
)

//go:embed assets/icons/*-64.png
var iconAssets embed.FS

type StatusIcon string

const (
	STATUS_ICON_APPLIED     StatusIcon = "applied"
	STATUS_ICON_NOT_APPLIED StatusIcon = "not-applied"
	STATUS_ICON_INCOMPLETE  StatusIcon = "incomplete"
	STATUS_ICON_EXTERNAL    StatusIcon = "external"
	STATUS_ICON_UNKNOWN     StatusIcon = "unknown"
	STATUS_ICON_CONFLICT    StatusIcon = "conflict"
)

func loadStatusIcon(icon StatusIcon) (*walk.Bitmap, error) {
	assetPath := path.Join("assets", "icons", string(icon)+"-64.png")

	data, err := iconAssets.ReadFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("load status icon %q: %w", icon, err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode status icon %q: %w", icon, err)
	}

	return walk.NewBitmapFromImageForDPI(img, 192)
}

func loadStatusIcons() (map[StatusIcon]*walk.Bitmap, error) {
	icons := make(map[StatusIcon]*walk.Bitmap)

	for _, icon := range []StatusIcon{
		STATUS_ICON_APPLIED,
		STATUS_ICON_NOT_APPLIED,
		STATUS_ICON_INCOMPLETE,
		STATUS_ICON_EXTERNAL,
		STATUS_ICON_UNKNOWN,
		STATUS_ICON_CONFLICT,
	} {
		bitmap, err := loadStatusIcon(icon)
		if err != nil {
			disposeStatusIcons(icons)
			return nil, err
		}

		icons[icon] = bitmap
	}

	return icons, nil
}

func disposeStatusIcons(icons map[StatusIcon]*walk.Bitmap) {
	for _, bitmap := range icons {
		bitmap.Dispose()
	}
}
