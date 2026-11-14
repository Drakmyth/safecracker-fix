package installer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Drakmyth/safecracker-fix/internal/dgvoodoo"
)

type GraphicsShimState int

const (
	GRAPHICS_SHIM_STATE_UNKNOWN GraphicsShimState = iota
	GRAPHICS_SHIM_STATE_NOT_INSTALLED
	GRAPHICS_SHIM_STATE_MANAGED
	GRAPHICS_SHIM_STATE_EXTERNAL
	GRAPHICS_SHIM_STATE_INCOMPLETE
	GRAPHICS_SHIM_STATE_CONFLICT
)

func (s GraphicsShimState) String() string {
	switch s {
	case GRAPHICS_SHIM_STATE_NOT_INSTALLED:
		return "Not installed"
	case GRAPHICS_SHIM_STATE_MANAGED:
		return "Managed installation"
	case GRAPHICS_SHIM_STATE_EXTERNAL:
		return "External installation"
	case GRAPHICS_SHIM_STATE_INCOMPLETE:
		return "Incomplete installation"
	case GRAPHICS_SHIM_STATE_CONFLICT:
		return "Conflicting installation"
	default:
		return "Unknown"
	}
}

type GraphicsShimInspection struct {
	State         GraphicsShimState
	DLLExists     bool
	DLLMatches    bool
	ConfigExists  bool
	ConfigManaged bool
	ConfigMatches bool
}

func (i *Installation) InspectGraphicsShim() (*GraphicsShimInspection, error) {
	bundle, err := dgvoodoo.Load()
	if err != nil {
		return nil, err
	}

	return inspectGraphicsShim(i.Directory, bundle)
}

func inspectGraphicsShim(directory string, bundle *dgvoodoo.Bundle) (*GraphicsShimInspection, error) {
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

	result := &GraphicsShimInspection{
		DLLExists:     dllExists,
		DLLMatches:    dllExists && bytes.Equal(dll, bundle.D3D9),
		ConfigExists:  configExists,
		ConfigManaged: configExists && dgvoodoo.IsManagedConfig(config),
		ConfigMatches: configExists && bytes.Equal(config, bundle.Config),
	}

	switch {
	case !result.DLLExists && !result.ConfigExists:
		result.State = GRAPHICS_SHIM_STATE_NOT_INSTALLED
	case result.DLLExists && !result.DLLMatches:
		result.State = GRAPHICS_SHIM_STATE_CONFLICT
	case result.DLLMatches && result.ConfigManaged:
		result.State = GRAPHICS_SHIM_STATE_MANAGED
	case result.DLLMatches:
		result.State = GRAPHICS_SHIM_STATE_EXTERNAL
	case result.ConfigManaged:
		result.State = GRAPHICS_SHIM_STATE_INCOMPLETE
	default:
		result.State = GRAPHICS_SHIM_STATE_CONFLICT
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
