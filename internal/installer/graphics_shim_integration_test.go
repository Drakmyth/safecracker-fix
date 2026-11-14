//go:build integration

package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Drakmyth/safecracker-fix/internal/dgvoodoo"
	"github.com/Drakmyth/safecracker-fix/internal/safecracker"
)

func TestGraphicsShimIntegration(t *testing.T) {
	fixturePath := filepath.Join("..", "safecracker", "testdata", "v101_steam", EXECUTABLE_NAME)

	original, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}

	directory := t.TempDir()
	executablePath := filepath.Join(directory, EXECUTABLE_NAME)

	if err := os.WriteFile(executablePath, original, 0600); err != nil {
		t.Fatal(err)
	}

	bundle, err := dgvoodoo.Load()
	if err != nil {
		t.Fatal(err)
	}

	dllPath := filepath.Join(directory, dgvoodoo.D3D9_FILENAME)
	configPath := filepath.Join(directory, dgvoodoo.CONFIG_FILENAME)

	installation, err := Inspect(directory)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the clean installation.
	graphics, err := installation.InspectGraphicsShim()
	if err != nil {
		t.Fatal(err)
	}

	if graphics.State != GRAPHICS_SHIM_STATE_NOT_INSTALLED {
		t.Fatalf("unexpected initial graphics shim state: %s", graphics.State)
	}

	// Install the actual bundled graphics shim.
	graphics, err = installation.InstallGraphicsShim()
	if err != nil {
		t.Fatal(err)
	}

	if graphics.State != GRAPHICS_SHIM_STATE_MANAGED {
		t.Fatalf("unexpected state after apply: %s", graphics.State)
	}

	assertFileContents(t, dllPath, bundle.D3D9)
	assertFileContents(t, configPath, bundle.Config)
	assertFileContents(t, executablePath, original)

	// Installing the shim must not create an executable backup.
	assertFileContents(t, installation.BackupPath(), nil)

	// Apply the performance fix independently.
	installation, err = installation.Apply(safecracker.PATCH_FPU_PRESERVE)
	if err != nil {
		t.Fatal(err)
	}

	patched, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(patched, original) {
		t.Fatal("performance fix did not modify executable")
	}

	assertFileContents(t, dllPath, bundle.D3D9)
	assertFileContents(t, configPath, bundle.Config)
	assertFileContents(t, installation.BackupPath(), original)

	// Remove the shim without reversing the performance fix.
	graphics, err = installation.RemoveGraphicsShim()
	if err != nil {
		t.Fatal(err)
	}

	if graphics.State != GRAPHICS_SHIM_STATE_NOT_INSTALLED {
		t.Fatalf("unexpected state after restore: %s", graphics.State)
	}

	assertFileContents(t, dllPath, nil)
	assertFileContents(t, configPath, nil)
	assertFileContents(t, executablePath, patched)

	// Reinstall the shim while the executable remains patched.
	graphics, err = installation.InstallGraphicsShim()
	if err != nil {
		t.Fatal(err)
	}

	if graphics.State != GRAPHICS_SHIM_STATE_MANAGED {
		t.Fatalf("unexpected state after reinstall: %s", graphics.State)
	}

	// Reverse the performance fix without removing the shim.
	installation, err = installation.Reverse(safecracker.PATCH_FPU_PRESERVE)
	if err != nil {
		t.Fatal(err)
	}

	assertFileContents(t, executablePath, original)
	assertFileContents(t, dllPath, bundle.D3D9)
	assertFileContents(t, configPath, bundle.Config)

	// Restore the graphics shim again.
	graphics, err = installation.RemoveGraphicsShim()
	if err != nil {
		t.Fatal(err)
	}

	if graphics.State != GRAPHICS_SHIM_STATE_NOT_INSTALLED {
		t.Fatalf("unexpected final state: %s", graphics.State)
	}

	// Verify the original game files and retained backup.
	assertFileContents(t, executablePath, original)
	assertFileContents(t, installation.BackupPath(), original)
	assertFileContents(t, dllPath, nil)
	assertFileContents(t, configPath, nil)

	assertNoTemporaryFiles(t, directory)

	// The integration fixture must remain untouched.
	assertFileContents(t, fixturePath, original)
}
