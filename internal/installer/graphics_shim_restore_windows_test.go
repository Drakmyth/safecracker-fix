package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Drakmyth/safecracker-fix/internal/dgvoodoo"
)

func TestRestoreGraphicsShimLockedDLL(t *testing.T) {
	directory := t.TempDir()

	bundle := &dgvoodoo.Bundle{
		D3D9:   []byte{0x01, 0x02, 0x03},
		Config: []byte(dgvoodoo.CONFIG_OWNERSHIP_MARKER + "\n\nVersion = 0x287\n"),
	}

	dllPath := filepath.Join(directory, dgvoodoo.D3D9_FILENAME)
	configPath := filepath.Join(directory, dgvoodoo.CONFIG_FILENAME)

	// Install the shim.
	if _, err := applyGraphicsShim(directory, bundle); err != nil {
		t.Fatal(err)
	}

	// Hold the DLL open without delete sharing.
	file, err := os.Open(dllPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Restoration should fail while the DLL is locked.
	got, err := restoreGraphicsShim(directory, bundle)
	if err == nil {
		t.Fatal("expected restoration to fail with locked DLL")
	}

	if got != nil {
		t.Fatal("expected nil inspection on error")
	}

	// Both files must remain intact.
	assertFileContents(t, dllPath, bundle.D3D9)
	assertFileContents(t, configPath, bundle.Config)

	inspection, err := inspectGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if inspection.State != GRAPHICS_SHIM_STATE_MANAGED {
		t.Fatalf("unexpected state after failed restore: %s", inspection.State)
	}

	// Release the lock and retry.
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	got, err = restoreGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if got.State != GRAPHICS_SHIM_STATE_NOT_INSTALLED {
		t.Fatalf("unexpected state after retry: %s", got.State)
	}

	assertFileContents(t, dllPath, nil)
	assertFileContents(t, configPath, nil)
}

func TestRestoreGraphicsShimLockedConfig(t *testing.T) {
	directory := t.TempDir()

	bundle := &dgvoodoo.Bundle{
		D3D9:   []byte{0x01, 0x02, 0x03},
		Config: []byte(dgvoodoo.CONFIG_OWNERSHIP_MARKER + "\n\nVersion = 0x287\n"),
	}

	dllPath := filepath.Join(directory, dgvoodoo.D3D9_FILENAME)
	configPath := filepath.Join(directory, dgvoodoo.CONFIG_FILENAME)

	if _, err := applyGraphicsShim(directory, bundle); err != nil {
		t.Fatal(err)
	}

	// Prevent configuration removal.
	file, err := os.Open(configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	got, err := restoreGraphicsShim(directory, bundle)
	if err == nil {
		t.Fatal("expected restoration to fail with locked config")
	}

	if got != nil {
		t.Fatal("expected nil inspection on error")
	}

	// The DLL was removed, but the configuration remains.
	assertFileContents(t, dllPath, nil)
	assertFileContents(t, configPath, bundle.Config)

	inspection, err := inspectGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if inspection.State != GRAPHICS_SHIM_STATE_INCOMPLETE {
		t.Fatalf("unexpected state after failed restore: %s", inspection.State)
	}

	// Release the lock and finish restoration.
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	got, err = restoreGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if got.State != GRAPHICS_SHIM_STATE_NOT_INSTALLED {
		t.Fatalf("unexpected state after retry: %s", got.State)
	}

	assertFileContents(t, dllPath, nil)
	assertFileContents(t, configPath, nil)
}
