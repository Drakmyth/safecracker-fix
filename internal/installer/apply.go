package installer

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Drakmyth/safecracker-fix/internal/patcher"
)

// Apply installs the specified patch on disk and return the updated installation.
func (i Installation) Apply(id patcher.PatchID) (*Installation, error) {
	return i.modifyExecutable(id, i.Target.Apply)
}

// Reverse removes the specified patch from disk and returns the updated installation.
func (i Installation) Reverse(id patcher.PatchID) (*Installation, error) {
	return i.modifyExecutable(id, i.Target.Reverse)
}

// modifyExecutable verifies, modifies, backs up, and replaces the installed executable.
func (i Installation) modifyExecutable(id patcher.PatchID, modify func([]byte, patcher.PatchID) ([]byte, error)) (*Installation, error) {
	// Always read the current executable. The installation inspection may no longer reflect its actual contents.
	current, err := os.ReadFile(i.ExecutablePath)
	if err != nil {
		return nil, fmt.Errorf("read game executable: %w", err)
	}

	modified, err := modify(current, id)
	if err != nil {
		return nil, fmt.Errorf("modify executable: %w", err)
	}

	// Avoid filesystem writes when the requested patch is already in the desired state.
	if bytes.Equal(current, modified) {
		return i.refresh()
	}

	if err := i.EnsureBackup(); err != nil {
		return nil, fmt.Errorf("ensure backup: %w", err)
	}

	// Detect changes that occurred while preparing the backup.
	latest, err := os.ReadFile(i.ExecutablePath)
	if err != nil {
		return nil, fmt.Errorf("reread game executable: %w", err)
	}

	if !bytes.Equal(current, latest) {
		return nil, fmt.Errorf("game executable changed during patch preparation")
	}

	if err := replaceFile(i.ExecutablePath, modified); err != nil {
		return nil, fmt.Errorf("replace game executable: %w", err)
	}

	// Verify the actual bytes written to disk.
	written, err := os.ReadFile(i.ExecutablePath)
	if err != nil {
		return nil, fmt.Errorf("verify game executable: %w", err)
	}

	if !bytes.Equal(written, modified) {
		return nil, fmt.Errorf("installed executable differs from expected result")
	}

	return i.refresh()
}
