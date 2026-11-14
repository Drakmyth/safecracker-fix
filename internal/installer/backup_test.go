package installer

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"github.com/Drakmyth/safecracker-fix/internal/patcher"
)

func TestEnsureBackup(t *testing.T) {
	const (
		TARGET_ID patcher.TargetID = "test-target"
		PATCH_A   patcher.PatchID  = "a"
		PATCH_B   patcher.PatchID  = "b"
	)

	original := []byte{0x00, 0x00, 0x00, 0x00}
	patched := []byte{0xAA, 0x00, 0xAA, 0xBB}
	partial := []byte{0xAA, 0x00, 0x00, 0x00}
	invalid := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	target := patcher.Target{
		ID:           TARGET_ID,
		Size:         int64(len(original)),
		OriginalHash: sha256.Sum256(original),
		Patches: []patcher.Patch{
			{
				ID: PATCH_A,
				Edits: []patcher.Edit{
					{
						Offset:      0,
						Original:    []byte{0x00},
						Replacement: []byte{0xAA},
					},
					{
						Offset:      2,
						Original:    []byte{0x00},
						Replacement: []byte{0xAA},
					},
				},
			},
			{
				ID: PATCH_B,
				Edits: []patcher.Edit{
					{
						Offset:      3,
						Original:    []byte{0x00},
						Replacement: []byte{0xBB},
					},
				},
			},
		},
	}

	tests := []struct {
		name           string
		executable     []byte
		existingBackup []byte
		wantBackup     []byte
		wantErr        bool
	}{
		{
			name:       "backup original executable",
			executable: original,
			wantBackup: original,
		},
		{
			name:       "backup fully patched executable",
			executable: patched,
			wantBackup: original,
		},
		{
			name:       "backup partially patched executable",
			executable: partial,
			wantBackup: original,
		},
		{
			name:           "preserve valid backup",
			executable:     patched,
			existingBackup: original,
			wantBackup:     original,
		},
		{
			name:           "reject invalid existing backup",
			executable:     original,
			existingBackup: invalid,
			wantBackup:     invalid,
			wantErr:        true,
		},
		{
			name:           "reject patched existing backup",
			executable:     original,
			existingBackup: patched,
			wantBackup:     patched,
			wantErr:        true,
		},
		{
			name:           "reject empty existing backup",
			executable:     original,
			existingBackup: []byte{},
			wantBackup:     []byte{},
			wantErr:        true,
		},
		{
			name:       "reject unsupported executable",
			executable: invalid,
			wantErr:    true,
		},
		{
			name:    "reject missing executable",
			wantErr: true,
		},
		{
			name:           "valid backup with missing executable",
			existingBackup: original,
			wantBackup:     original,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()

			installation := Installation{
				Directory:      directory,
				ExecutablePath: filepath.Join(directory, EXECUTABLE_NAME),
				Target:         target,
			}

			backupPath := installation.BackupPath()

			if tt.executable != nil {
				if err := os.WriteFile(installation.ExecutablePath, tt.executable, 0600); err != nil {
					t.Fatal(err)
				}
			}

			if tt.existingBackup != nil {
				if err := os.WriteFile(backupPath, tt.existingBackup, 0600); err != nil {
					t.Fatal(err)
				}
			}

			err := installation.EnsureBackup()

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			// Backup creation must never modify or create
			// the installed executable.
			assertFileContents(t, installation.ExecutablePath, tt.executable)

			// Verify the expected backup, including
			// preservation of invalid existing backups.
			assertFileContents(t, backupPath, tt.wantBackup)
		})
	}
}

// assertFileContents verifies a file's exact contents.
// A nil expected slice indicates that the file must not exist.
func assertFileContents(t *testing.T, path string, expected []byte) {
	t.Helper()

	actual, err := os.ReadFile(path)

	if expected == nil {
		if !os.IsNotExist(err) {
			t.Errorf("expected %q not to exist, got error %v", path, err)
		}
		return
	}

	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}

	if !bytes.Equal(actual, expected) {
		t.Errorf("%q contents = %x, want %x", path, actual, expected)
	}
}

func TestBackupPath(t *testing.T) {
	installation := Installation{
		ExecutablePath: filepath.Join("game", "Safecracker.exe"),
	}

	want := filepath.Join("game", "Safecracker.exe.safecracker-fix.bak")

	if got := installation.BackupPath(); got != want {
		t.Errorf("BackupPath() = %q, want %q", got, want)
	}
}

func TestWriteBackupDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.bak")

	original := []byte{0x00, 0x01, 0x02}
	replacement := []byte{0xAA, 0xBB, 0xCC}

	if err := writeBackup(path, original); err != nil {
		t.Fatalf("create backup: %v", err)
	}

	assertFileContents(t, path, original)

	if err := writeBackup(path, replacement); err == nil {
		t.Fatal("expected error when backup already exists")
	}

	assertFileContents(t, path, original)
}
