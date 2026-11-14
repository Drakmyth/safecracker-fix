package installer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Drakmyth/safecracker-fix/internal/dgvoodoo"
)

type GraphicsState int

const (
	GRAPHICS_STATE_UNKNOWN GraphicsState = iota
	GRAPHICS_STATE_NOT_INSTALLED
	GRAPHICS_STATE_MANAGED
	GRAPHICS_STATE_EXTERNAL
	GRAPHICS_STATE_INCOMPLETE
	GRAPHICS_STATE_CONFLICT
)

func (s GraphicsState) String() string {
	switch s {
	case GRAPHICS_STATE_NOT_INSTALLED:
		return "Not installed"
	case GRAPHICS_STATE_MANAGED:
		return "Managed installation"
	case GRAPHICS_STATE_EXTERNAL:
		return "External installation"
	case GRAPHICS_STATE_INCOMPLETE:
		return "Incomplete installation"
	case GRAPHICS_STATE_CONFLICT:
		return "Conflicting installation"
	default:
		return "Unknown"
	}
}

type GraphicsInspection struct {
	State         GraphicsState
	DLLExists     bool
	DLLMatches    bool
	ConfigExists  bool
	ConfigManaged bool
	ConfigMatches bool
}

func (i *Installation) InspectGraphics() (*GraphicsInspection, error) {
	bundle, err := dgvoodoo.Load()
	if err != nil {
		return nil, err
	}

	return inspectGraphics(i.Directory, bundle)
}

func inspectGraphics(directory string, bundle *dgvoodoo.Bundle) (*GraphicsInspection, error) {
	dllPath := filepath.Join(directory, dgvoodoo.D3D9_FILENAME)
	configPath := filepath.Join(directory, dgvoodoo.CONFIG_FILENAME)

	dll, dllExists, err := readOptionalFile(dllPath)
	if err != nil {
		return nil, err
	}

	config, configExists, err := readOptionalFile(configPath)
	if err != nil {
		return nil, err
	}

	result := &GraphicsInspection{
		DLLExists:     dllExists,
		DLLMatches:    dllExists && bytes.Equal(dll, bundle.D3D9),
		ConfigExists:  configExists,
		ConfigManaged: configExists && dgvoodoo.IsManagedConfig(config),
		ConfigMatches: configExists && bytes.Equal(config, bundle.Config),
	}

	switch {
	case !result.DLLExists && !result.ConfigExists:
		result.State = GRAPHICS_STATE_NOT_INSTALLED
	case result.DLLExists && !result.DLLMatches:
		result.State = GRAPHICS_STATE_CONFLICT
	case result.DLLMatches && result.ConfigManaged:
		result.State = GRAPHICS_STATE_MANAGED
	case result.DLLMatches:
		result.State = GRAPHICS_STATE_EXTERNAL
	case result.ConfigManaged:
		result.State = GRAPHICS_STATE_INCOMPLETE
	default:
		result.State = GRAPHICS_STATE_CONFLICT
	}

	return result, nil
}

func readOptionalFile(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)

	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("read %q: %w", path, err)
	}

	return data, true, nil
}
