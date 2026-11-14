package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Drakmyth/safecracker-fix/internal/patcher"
	"github.com/Drakmyth/safecracker-fix/internal/safecracker"
)

const EXECUTABLE_NAME = "Safecracker.exe"

type Installation struct {
	Directory      string
	ExecutablePath string
	Target         patcher.Target
	Inspection     *patcher.Inspection
}

// Inspect identifies a Safecracker installation and determines the state of its supported patches.
// It does not modify any files.
func Inspect(directory string) (*Installation, error) {
	return inspect(directory, safecracker.SupportedTargets())
}

func inspect(directory string, targets []patcher.Target) (*Installation, error) {
	if directory == "" {
		return nil, fmt.Errorf("installation directory is empty")
	}

	directory, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve installation directory: %w", err)
	}

	executablePath := filepath.Join(directory, EXECUTABLE_NAME)

	data, err := os.ReadFile(executablePath)
	if err != nil {
		return nil, fmt.Errorf("read game executable: %w", err)
	}

	for _, target := range targets {
		inspection, err := target.Inspect(data)
		if err != nil {
			continue
		}

		return &Installation{
			Directory:      directory,
			ExecutablePath: executablePath,
			Target:         target,
			Inspection:     inspection,
		}, nil
	}

	return nil, fmt.Errorf("unsupported Safecraacker executable: %s", executablePath)
}

// refresh returns an updated inspection of this installation using its previously identified target.
func (i Installation) refresh() (*Installation, error) {
	return inspect(i.Directory, []patcher.Target{i.Target})
}
