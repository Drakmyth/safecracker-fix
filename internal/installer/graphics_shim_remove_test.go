package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Drakmyth/safecracker-fix/internal/dgvoodoo"
)

func TestRestoreGraphicsShim(t *testing.T) {
	bundle := &dgvoodoo.Bundle{
		D3D9:   []byte{0x01, 0x02, 0x03},
		Config: []byte(dgvoodoo.CONFIG_OWNERSHIP_MARKER + "\n\nVersion = 0x287\n"),
	}

	otherDLL := []byte{0xAA, 0xBB, 0xCC}
	otherConfig := []byte("Version = 0x287\n")

	modifiedConfig := append(bytes.Clone(bundle.Config), []byte("\n[DirectX]\nFastVideoMemoryAccess = true\n")...)

	tests := []struct {
		name       string
		dll        []byte
		config     []byte
		wantDLL    []byte
		wantConfig []byte
		wantState  GraphicsShimState
		wantErr    bool
	}{
		{
			name:      "not installed",
			wantState: GRAPHICS_SHIM_STATE_NOT_INSTALLED,
		},
		{
			name:      "managed installation",
			dll:       bundle.D3D9,
			config:    bundle.Config,
			wantState: GRAPHICS_SHIM_STATE_NOT_INSTALLED,
		},
		{
			name:       "managed installation with modified config",
			dll:        bundle.D3D9,
			config:     modifiedConfig,
			wantConfig: modifiedConfig,
			wantState:  GRAPHICS_SHIM_STATE_INCOMPLETE,
		},
		{
			name:      "incomplete managed installation",
			config:    bundle.Config,
			wantState: GRAPHICS_SHIM_STATE_NOT_INSTALLED,
		},
		{
			name:       "incomplete with modified config",
			config:     modifiedConfig,
			wantConfig: modifiedConfig,
			wantState:  GRAPHICS_SHIM_STATE_INCOMPLETE,
		},
		{
			name:      "external installation without config",
			dll:       bundle.D3D9,
			wantDLL:   bundle.D3D9,
			wantState: GRAPHICS_SHIM_STATE_EXTERNAL,
		},
		{
			name:       "external installation with config",
			dll:        bundle.D3D9,
			config:     otherConfig,
			wantDLL:    bundle.D3D9,
			wantConfig: otherConfig,
			wantState:  GRAPHICS_SHIM_STATE_EXTERNAL,
		},
		{
			name:    "conflicting DLL",
			dll:     otherDLL,
			wantDLL: otherDLL,
			wantErr: true,
		},
		{
			name:       "conflicting config",
			config:     otherConfig,
			wantConfig: otherConfig,
			wantErr:    true,
		},
		{
			name:       "conflicting DLL and external config",
			dll:        otherDLL,
			config:     otherConfig,
			wantDLL:    otherDLL,
			wantConfig: otherConfig,
			wantErr:    true,
		},
		{
			name:       "conflicting DLL and managed config",
			dll:        otherDLL,
			config:     bundle.Config,
			wantDLL:    otherDLL,
			wantConfig: bundle.Config,
			wantErr:    true,
		},
		{
			name:       "conflicting DLL and modified managed config",
			dll:        otherDLL,
			config:     modifiedConfig,
			wantDLL:    otherDLL,
			wantConfig: modifiedConfig,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()

			dllPath := filepath.Join(directory, dgvoodoo.D3D9_FILENAME)
			configPath := filepath.Join(directory, dgvoodoo.CONFIG_FILENAME)

			if tt.dll != nil {
				if err := os.WriteFile(dllPath, tt.dll, 0600); err != nil {
					t.Fatal(err)
				}
			}

			if tt.config != nil {
				if err := os.WriteFile(configPath, tt.config, 0600); err != nil {
					t.Fatal(err)
				}
			}

			got, err := removeGraphicsShim(directory, bundle)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected restoration error")
				}
				if got != nil {
					t.Fatal("expected nil inspection on error")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if got == nil {
					t.Fatal("expected graphics shim inspection")
				}
				if got.State != tt.wantState {
					t.Errorf("state = %s, want %s", got.State, tt.wantState)
				}
			}

			assertFileContents(t, dllPath, tt.wantDLL)
			assertFileContents(t, configPath, tt.wantConfig)
		})
	}
}

func TestGraphicsShimLifecycle(t *testing.T) {
	directory := t.TempDir()

	bundle := &dgvoodoo.Bundle{
		D3D9:   []byte{0x01, 0x02, 0x03},
		Config: []byte(dgvoodoo.CONFIG_OWNERSHIP_MARKER + "\n\nVersion = 0x287\n"),
	}

	dllPath := filepath.Join(
		directory, dgvoodoo.D3D9_FILENAME,
	)
	configPath := filepath.Join(
		directory, dgvoodoo.CONFIG_FILENAME,
	)

	// Install the shim.
	got, err := installGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if got.State != GRAPHICS_SHIM_STATE_MANAGED {
		t.Fatalf("unexpected state after apply: %s", got.State)
	}

	// Simulate the user customizing the configuration.
	modifiedConfig := append(bytes.Clone(bundle.Config), []byte("\n[DirectX]\nFastVideoMemoryAccess = true\n")...)

	if err := os.WriteFile(configPath, modifiedConfig, 0600); err != nil {
		t.Fatal(err)
	}

	// Restore removes the DLL but preserves the configuration.
	got, err = removeGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if got.State != GRAPHICS_SHIM_STATE_INCOMPLETE {
		t.Fatalf("unexpected state after restore: %s", got.State)
	}

	assertFileContents(t, dllPath, nil)
	assertFileContents(t, configPath, modifiedConfig)

	// Repeated restoration must preserve the configuration.
	got, err = removeGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if got.State != GRAPHICS_SHIM_STATE_INCOMPLETE {
		t.Fatalf("unexpected state after repeated restore: %s", got.State)
	}

	assertFileContents(t, dllPath, nil)
	assertFileContents(t, configPath, modifiedConfig)

	// Reinstall the DLL using the retained configuration.
	got, err = installGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if got.State != GRAPHICS_SHIM_STATE_MANAGED {
		t.Fatalf("unexpected state after reinstall: %s", got.State)
	}

	if !got.ConfigManaged || got.ConfigMatches {
		t.Errorf("unexpected configuration state: %+v", got)
	}

	assertFileContents(t, dllPath, bundle.D3D9)
	assertFileContents(t, configPath, modifiedConfig)

	// Return the configuration to its original contents.
	if err := os.WriteFile(configPath, bundle.Config, 0600); err != nil {
		t.Fatal(err)
	}

	// Restore now removes both files.
	got, err = removeGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if got.State != GRAPHICS_SHIM_STATE_NOT_INSTALLED {
		t.Fatalf("unexpected final state: %s", got.State)
	}

	assertFileContents(t, dllPath, nil)
	assertFileContents(t, configPath, nil)
}
