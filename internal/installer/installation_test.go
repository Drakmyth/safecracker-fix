package installer

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"github.com/Drakmyth/safecracker-fix/internal/patcher"
)

func TestInspect(t *testing.T) {
	const (
		TARGET_ID patcher.TargetID = "test-target"
		PATCH_ID  patcher.PatchID  = "test-patch"
	)

	original := []byte{0x00, 0x00, 0x00, 0x00}

	target := patcher.Target{
		ID:           TARGET_ID,
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
					{
						Offset:      2,
						Original:    []byte{0x00},
						Replacement: []byte{0xAA},
					},
				},
			},
		},
	}

	tests := []struct {
		name    string
		data    []byte
		state   patcher.PatchState
		wantErr bool
	}{
		{
			name:  "original executable",
			data:  original,
			state: patcher.PATCH_STATE_UNAPPLIED,
		},
		{
			name:  "patched executable",
			data:  []byte{0xAA, 0x00, 0xAA, 0x00},
			state: patcher.PATCH_STATE_APPLIED,
		},
		{
			name:  "partially patched executable",
			data:  []byte{0xAA, 0x00, 0x00, 0x00},
			state: patcher.PATCH_STATE_PARTIAL,
		},
		{
			name:    "missing executable",
			wantErr: true,
		},
		{
			name:    "empty executable",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "unsupported executable",
			data:    []byte{0xFF, 0xFF, 0xFF, 0xFF},
			wantErr: true,
		},
		{
			name:    "incorrect executable size",
			data:    []byte{0x00},
			wantErr: true,
		},
		{
			name:    "unrelated modification",
			data:    []byte{0x00, 0xFF, 0x00, 0x00},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()
			executablePath := filepath.Join(directory, EXECUTABLE_NAME)

			if tt.data != nil {
				if err := os.WriteFile(executablePath, tt.data, 0600); err != nil {
					t.Fatal(err)
				}
			}

			result, err := inspect(directory, []patcher.Target{target})

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			// Inspection must never modify the executable.
			if tt.data != nil {
				after, err := os.ReadFile(executablePath)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(after, tt.data) {
					t.Error("inspection modified executable")
				}
			}

			if tt.wantErr {
				if result != nil {
					t.Error("inspection returned result on error")
				}
				return
			}

			if result == nil {
				t.Fatal("inspection returned nil result")
			}

			if result.Directory != directory {
				t.Errorf("directory = %q, want %q", result.Directory, directory)
			}

			if result.ExecutablePath != executablePath {
				t.Errorf("executable path = %q, want %q", result.ExecutablePath, executablePath)
			}

			if result.Target.ID != TARGET_ID {
				t.Errorf("target ID = %q, want %q", result.Target.ID, TARGET_ID)
			}

			if result.Inspection == nil {
				t.Fatal("missing inspection")
			}

			state, ok := result.Inspection.Patches[PATCH_ID]
			if !ok {
				t.Fatal("missing patch state")
			}

			if state != tt.state {
				t.Errorf("patch state = %v, want %v", state, tt.state)
			}
		})
	}
}

func TestInspectEmptyDirectory(t *testing.T) {
	result, err := Inspect("")

	if err == nil {
		t.Fatal("expected error for empty directory")
	}

	if result != nil {
		t.Error("inspection returned result on error")
	}
}

func TestInspectMissingDirectory(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "missing")

	result, err := Inspect(directory)

	if err == nil {
		t.Fatal("expected error for missing directory")
	}

	if result != nil {
		t.Error("inspection returned result on error")
	}
}
