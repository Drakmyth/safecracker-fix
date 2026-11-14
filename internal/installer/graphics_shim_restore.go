package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Drakmyth/safecracker-fix/internal/dgvoodoo"
)

func (i *Installation) RestoreGraphicsShim() (*GraphicsShimInspection, error) {
	bundle, err := dgvoodoo.Load()
	if err != nil {
		return nil, err
	}

	return restoreGraphicsShim(i.Directory, bundle)
}

func restoreGraphicsShim(directory string, bundle *dgvoodoo.Bundle) (*GraphicsShimInspection, error) {
	inspection, err := inspectGraphicsShim(directory, bundle)
	if err != nil {
		return nil, err
	}

	dllPath := filepath.Join(directory, dgvoodoo.D3D9_FILENAME)
	configPath := filepath.Join(directory, dgvoodoo.CONFIG_FILENAME)

	switch inspection.State {
	case GRAPHICS_SHIM_STATE_NOT_INSTALLED, GRAPHICS_SHIM_STATE_EXTERNAL:
		return inspection, nil
	case GRAPHICS_SHIM_STATE_MANAGED, GRAPHICS_SHIM_STATE_INCOMPLETE:
		if inspection.DLLMatches {
			if err := os.Remove(dllPath); err != nil {
				return nil, fmt.Errorf("remove graphics shim DLL: %w", err)
			}
		}

		if inspection.ConfigMatches {
			if err := os.Remove(configPath); err != nil {
				return nil, fmt.Errorf("remove graphics shim configuration: %w", err)
			}
		}
	case GRAPHICS_SHIM_STATE_CONFLICT:
		return nil, fmt.Errorf("graphics shim installation conflicts with existing files")
	default:
		return nil, fmt.Errorf("unexpected graphics shim state: %s", inspection.State)
	}

	return inspectGraphicsShim(directory, bundle)
}
