package installer

import (
	"fmt"
	"os"
)

func writeNewFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}

	success := false
	defer func() {
		file.Close()

		if !success {
			os.Remove(path)
		}
	}()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write %q: %w", path, err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync %q: %w", path, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close %q: %w", path, err)
	}

	success = true

	return nil
}
