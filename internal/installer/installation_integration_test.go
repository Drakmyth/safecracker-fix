//go:build integration

package installer

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Drakmyth/safecracker-fix/internal/patcher"
	"github.com/Drakmyth/safecracker-fix/internal/safecracker"
)

func TestInstallationV101Steam(t *testing.T) {
	const EXPECTED_PATCHED_HASH = "3ce6c4c4525ddc9e500d7fa7fa03f900f56dcd0398f43bba938226c866a2d1a3"

	// Reuse the original executable fixture.
	fixturePath := filepath.Join("..", "safecracker", "testdata", "v101_steam", EXECUTABLE_NAME)

	original, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read integration fixture: %v", err)
	}

	// Create an isolated game installation.
	directory := t.TempDir()
	executablePath := filepath.Join(directory, EXECUTABLE_NAME)

	if err := os.WriteFile(executablePath, original, 0600); err != nil {
		t.Fatal(err)
	}

	// Identify the real executable using the public registry.
	installation, err := Inspect(directory)
	if err != nil {
		t.Fatalf("inspect original installation: %v", err)
	}

	patchID := safecracker.PATCH_FPU_PRESERVE

	assertPatchState(t, installation, patchID, patcher.PATCH_STATE_UNAPPLIED)

	// Apply the performance fix.
	installation, err = installation.Apply(patchID)
	if err != nil {
		t.Fatalf("apply performance fix: %v", err)
	}

	assertPatchState(t, installation, patchID, patcher.PATCH_STATE_APPLIED)

	// Verify the installed executable matches the known working patched executable.
	patched, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatal(err)
	}

	actualHash := fmt.Sprintf("%x", sha256.Sum256(patched))

	if actualHash != EXPECTED_PATCHED_HASH {
		t.Fatalf("patched hash = %s, want %s", actualHash, EXPECTED_PATCHED_HASH)
	}

	// Verify the backup contains the exact original.
	assertFileContents(t, installation.BackupPath(), original)

	// Reinspect independently to simulate restarting the installer application.
	installation, err = Inspect(directory)
	if err != nil {
		t.Fatalf("inspect patched installation: %v", err)
	}

	assertPatchState(t, installation, patchID, patcher.PATCH_STATE_APPLIED)

	// Reverse the performance fix.
	installation, err = installation.Reverse(patchID)
	if err != nil {
		t.Fatalf("reverse performance fix: %v", err)
	}

	assertPatchState(t, installation, patchID, patcher.PATCH_STATE_UNAPPLIED)

	// The restored executable must match the original byte-for-byte, and the backup must remain intact.
	assertFileContents(t, executablePath, original)
	assertFileContents(t, installation.BackupPath(), original)

	assertNoTemporaryFiles(t, directory)

	// Confirm the source fixture was never modified.
	assertFileContents(t, fixturePath, original)
}

func assertPatchState(t *testing.T, installation *Installation, id patcher.PatchID, expected patcher.PatchState) {
	t.Helper()

	if installation == nil || installation.Inspection == nil {
		t.Fatal("missing installation inspection")
	}

	actual, exists := installation.Inspection.Patches[id]

	if !exists {
		t.Fatalf("patch %q missing from inspection", id)
	}

	if actual != expected {
		t.Errorf("patch %q state = %v, want %v", id, actual, expected)
	}
}
