package installer

import (
	"crypto/sha256"
	"fmt"
	"os"
)

const BACKUP_SUFFIX = ".safecracker-fix.bak"

// BackupPath returns the path to the original executable backup.
func (i Installation) BackupPath() string {
	return i.ExecutablePath + BACKUP_SUFFIX
}

// EnsureBackup creates a backup of the original executable if one does not already exist.
// An existing backup must match the target's original hash.
// Invalid backups are never overwritten.
func (i Installation) EnsureBackup() error {
	backupPath := i.BackupPath()

	existing, err := os.ReadFile(backupPath)
	if err == nil {
		if sha256.Sum256(existing) != i.Target.OriginalHash {
			return fmt.Errorf("existing backup is invalid: %s", backupPath)
		}

		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("read backup: %w", err)
	}

	// Read the current executable rather than relying on the installation's inspection snapshot.
	data, err := os.ReadFile(i.ExecutablePath)
	if err != nil {
		return fmt.Errorf("read game executable: %w", err)
	}

	if _, err := i.Target.Inspect(data); err != nil {
		return fmt.Errorf("inspect game executable: %w", err)
	}

	// Reconstruct the original executable by reversing every supported patch.
	original := data

	for _, patch := range i.Target.Patches {
		original, err = i.Target.Reverse(original, patch.ID)
		if err != nil {
			return fmt.Errorf("reverse patch %q for backup: %w", patch.ID, err)
		}
	}

	// Independently verify the reconstructed original.
	if sha256.Sum256(original) != i.Target.OriginalHash {
		return fmt.Errorf("reconstructed original hash mismatch")
	}

	return writeBackup(backupPath, original)
}

// writeBackup creates a new backup without overwriting any existing file.
func writeBackup(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create backup: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(path)
		return fmt.Errorf("write backup: %w", err)
	}

	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(path)
		return fmt.Errorf("sync backup: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(path)
		return fmt.Errorf("close backup: %w", err)
	}

	return nil
}
