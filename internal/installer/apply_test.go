package installer

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"github.com/Drakmyth/safecracker-fix/internal/patcher"
)

func TestModifyExecutable(t *testing.T) {
	const (
		TARGET_ID patcher.TargetID = "test-target"
		PATCH_A   patcher.PatchID  = "a"
		PATCH_B   patcher.PatchID  = "b"
	)

	original := []byte{0x00, 0x00, 0x00, 0x00}
	appliedA := []byte{0xAA, 0x00, 0xAA, 0x00}
	appliedB := []byte{0x00, 0x00, 0x00, 0xBB}
	appliedBoth := []byte{0xAA, 0x00, 0xAA, 0xBB}
	partialA := []byte{0xAA, 0x00, 0x00, 0x00}
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
		patchID        patcher.PatchID
		reverse        bool
		wantExecutable []byte
		wantBackup     []byte
		wantErr        bool
	}{
		{
			name:           "apply patch",
			executable:     original,
			patchID:        PATCH_A,
			wantExecutable: appliedA,
			wantBackup:     original,
		},
		{
			name:           "reverse patch",
			executable:     appliedA,
			patchID:        PATCH_A,
			reverse:        true,
			wantExecutable: original,
			wantBackup:     original,
		},
		{
			name:           "complete partial patch",
			executable:     partialA,
			patchID:        PATCH_A,
			wantExecutable: appliedA,
			wantBackup:     original,
		},
		{
			name:           "reverse partial patch",
			executable:     partialA,
			patchID:        PATCH_A,
			reverse:        true,
			wantExecutable: original,
			wantBackup:     original,
		},
		{
			name:           "apply preserves other patch",
			executable:     appliedB,
			patchID:        PATCH_A,
			wantExecutable: appliedBoth,
			wantBackup:     original,
		},
		{
			name:           "reverse preserves other patch",
			executable:     appliedBoth,
			patchID:        PATCH_A,
			reverse:        true,
			wantExecutable: appliedB,
			wantBackup:     original,
		},
		{
			name:           "apply already applied patch",
			executable:     appliedA,
			patchID:        PATCH_A,
			wantExecutable: appliedA,
		},
		{
			name:           "reverse unapplied patch",
			executable:     original,
			patchID:        PATCH_A,
			reverse:        true,
			wantExecutable: original,
		},
		{
			name:           "reuse existing valid backup",
			executable:     original,
			existingBackup: original,
			patchID:        PATCH_A,
			wantExecutable: appliedA,
			wantBackup:     original,
		},
		{
			name:           "reject invalid backup",
			executable:     original,
			existingBackup: invalid,
			patchID:        PATCH_A,
			wantExecutable: original,
			wantBackup:     invalid,
			wantErr:        true,
		},
		{
			name:           "reject unknown patch",
			executable:     original,
			patchID:        "unknown",
			wantExecutable: original,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, EXECUTABLE_NAME)

			if err := os.WriteFile(path, tt.executable, 0600); err != nil {
				t.Fatal(err)
			}

			installation, err := inspect(directory, []patcher.Target{target})
			if err != nil {
				t.Fatal(err)
			}

			if tt.existingBackup != nil {
				if err := os.WriteFile(installation.BackupPath(), tt.existingBackup, 0600); err != nil {
					t.Fatal(err)
				}
			}

			var result *Installation

			if tt.reverse {
				result, err = installation.Reverse(tt.patchID)
			} else {
				result, err = installation.Apply(tt.patchID)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			assertFileContents(t, path, tt.wantExecutable)
			assertFileContents(t, installation.BackupPath(), tt.wantBackup)
			assertNoTemporaryFiles(t, directory)

			if tt.wantErr {
				if result != nil {
					t.Fatal("expected nil result on error")
				}
				return
			}

			if result == nil {
				t.Fatal("expected updated installation")
			}

			if result == installation {
				t.Fatal("expected a new installation snapshot")
			}

			if result.Directory != installation.Directory {
				t.Error("installation directory changed")
			}

			if result.ExecutablePath != path {
				t.Error("executable path changed")
			}

			if result.Target.ID != TARGET_ID {
				t.Error("target ID changed")
			}

			if result.Inspection == nil {
				t.Fatal("expected updated inspection")
			}

			expected, err := target.Inspect(tt.wantExecutable)
			if err != nil {
				t.Fatal(err)
			}

			for id, want := range expected.Patches {
				got, exists := result.Inspection.Patches[id]
				if !exists || got != want {
					t.Errorf("patch %q state = %v, want %v", id, got, want)
				}
			}
		})
	}
}

func TestModifyExecutableStaleInspection(t *testing.T) {
	const PATCH_ID patcher.PatchID = "test-patch"

	original := []byte{0x00, 0x00}
	patched := []byte{0xAA, 0x00}
	invalid := []byte{0x00, 0xFF}

	target := patcher.Target{
		ID:           "test-target",
		Size:         int64(len(original)),
		OriginalHash: sha256.Sum256(original),
		Patches: []patcher.Patch{
			{
				ID: PATCH_ID,
				Edits: []patcher.Edit{
					{
						Offset:      0,
						Original:    []byte{0x00},
						Replacement: []byte{0xAA},
					},
				},
			},
		},
	}

	tests := []struct {
		name           string
		changed        []byte
		wantExecutable []byte
		wantErr        bool
	}{
		{
			name:           "patch applied externally",
			changed:        patched,
			wantExecutable: patched,
		},
		{
			name:           "executable modified externally",
			changed:        invalid,
			wantExecutable: invalid,
			wantErr:        true,
		},
		{
			name:    "executable removed externally",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, EXECUTABLE_NAME)

			if err := os.WriteFile(path, original, 0600); err != nil {
				t.Fatal(err)
			}

			installation, err := inspect(directory, []patcher.Target{target})
			if err != nil {
				t.Fatal(err)
			}

			// Change the executable after inspection.
			if tt.changed == nil {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(path, tt.changed, 0600); err != nil {
					t.Fatal(err)
				}
			}

			result, err := installation.Apply(PATCH_ID)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			assertFileContents(t, path, tt.wantExecutable)
			assertFileContents(t, installation.BackupPath(), nil)
			assertNoTemporaryFiles(t, directory)

			// The original snapshot must remain unchanged.
			if got := installation.Inspection.Patches[PATCH_ID]; got != patcher.PATCH_STATE_UNAPPLIED {
				t.Errorf("original snapshot changed: %v", got)
			}

			if tt.wantErr {
				if result != nil {
					t.Fatal("expected nil result on error")
				}
				return
			}

			if result == nil || result.Inspection == nil {
				t.Fatal("expected updated inspection")
			}

			if got := result.Inspection.Patches[PATCH_ID]; got != patcher.PATCH_STATE_APPLIED {
				t.Errorf("updated patch state = %v", got)
			}
		})
	}
}
