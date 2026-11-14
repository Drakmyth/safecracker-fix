package installer

import (
	"fmt"
	"path/filepath"

	"github.com/Drakmyth/safecracker-fix/internal/dgvoodoo"
)

func (i *Installation) ApplyGraphicsShim() (*GraphicsShimInspection, error) {
	bundle, err := dgvoodoo.Load()
	if err != nil {
		return nil, err
	}

	return applyGraphicsShim(i.Directory, bundle)
}

func applyGraphicsShim(directory string, bundle *dgvoodoo.Bundle) (*GraphicsShimInspection, error) {
	inspection, err := inspectGraphicsShim(directory, bundle)
	if err != nil {
		return nil, err
	}

	dllPath := filepath.Join(directory, dgvoodoo.D3D9_FILENAME)
	configPath := filepath.Join(directory, dgvoodoo.CONFIG_FILENAME)

	switch inspection.State {
	case GRAPHICS_SHIM_STATE_MANAGED, GRAPHICS_SHIM_STATE_EXTERNAL:
		return inspection, nil
	case GRAPHICS_SHIM_STATE_NOT_INSTALLED:
		if err := writeNewFile(configPath, bundle.Config); err != nil {
			return nil, fmt.Errorf("install graphics shim configuration: %w", err)
		}

		if err := writeNewFile(dllPath, bundle.D3D9); err != nil {
			return nil, fmt.Errorf("repair graphics shim installation: %w", err)
		}
	case GRAPHICS_SHIM_STATE_INCOMPLETE:
		if err := writeNewFile(dllPath, bundle.D3D9); err != nil {
			return nil, fmt.Errorf("repair graphics shim installation: %w", err)
		}
	case GRAPHICS_SHIM_STATE_CONFLICT:
		return nil, fmt.Errorf("graphics shim installation conflicts with existing files")
	default:
		return nil, fmt.Errorf("unexpected graphics shim state: %s", inspection.State)
	}

	result, err := inspectGraphicsShim(directory, bundle)
	if err != nil {
		return nil, err
	}

	if result.State != GRAPHICS_SHIM_STATE_MANAGED {
		return nil, fmt.Errorf("graphics shim installation verification failed: %s", result.State)
	}

	return result, nil
}
