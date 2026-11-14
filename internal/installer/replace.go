package installer

import (
	"fmt"
	"os"
	"path/filepath"
)

// replaceFile replaces an existing file with the supplied contents using a temporary file in the same directory.
// The destination is not modified until the temporary file has been completely written, synced, and closed.
func replaceFile(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat destination: %w", err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("destination is not a regular file: %s", path)
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".safecracker-fix-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}

	tempPath := file.Name()

	defer func() {
		file.Close()
		os.Remove(tempPath)
	}()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write temporary file: %w", err)
	}

	if err := file.Chmod(info.Mode().Perm()); err != nil {
		return fmt.Errorf("set temporary file permissions: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync temporary file: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace destination: %w", err)
	}

	return nil
}
