package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Drakmyth/safecracker-fix/internal/dgvoodoo"
)

func TestApplyGraphicsShim(t *testing.T) {
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
			name:       "clean installation",
			wantDLL:    bundle.D3D9,
			wantConfig: bundle.Config,
			wantState:  GRAPHICS_SHIM_STATE_MANAGED,
		},
		{
			name:       "already installed",
			dll:        bundle.D3D9,
			config:     bundle.Config,
			wantDLL:    bundle.D3D9,
			wantConfig: bundle.Config,
			wantState:  GRAPHICS_SHIM_STATE_MANAGED,
		},
		{
			name:       "preserve modified managed config",
			dll:        bundle.D3D9,
			config:     modifiedConfig,
			wantDLL:    bundle.D3D9,
			wantConfig: modifiedConfig,
			wantState:  GRAPHICS_SHIM_STATE_MANAGED,
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
			name:       "repair missing DLL",
			config:     bundle.Config,
			wantDLL:    bundle.D3D9,
			wantConfig: bundle.Config,
			wantState:  GRAPHICS_SHIM_STATE_MANAGED,
		},
		{
			name:       "repair preserves modified config",
			config:     modifiedConfig,
			wantDLL:    bundle.D3D9,
			wantConfig: modifiedConfig,
			wantState:  GRAPHICS_SHIM_STATE_MANAGED,
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
			name:       "conflicting DLL and managed config",
			dll:        otherDLL,
			config:     bundle.Config,
			wantDLL:    otherDLL,
			wantConfig: bundle.Config,
			wantErr:    true,
		},
		{
			name:    "empty DLL",
			dll:     []byte{},
			wantDLL: []byte{},
			wantErr: true,
		},
		{
			name:       "empty config",
			config:     []byte{},
			wantConfig: []byte{},
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

			got, err := applyGraphicsShim(directory, bundle)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected installation error")
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

func TestApplyGraphicsShimIdempotent(t *testing.T) {
	directory := t.TempDir()

	bundle := &dgvoodoo.Bundle{
		D3D9:   []byte{0x01, 0x02, 0x03},
		Config: []byte(dgvoodoo.CONFIG_OWNERSHIP_MARKER + "\n\nVersion = 0x287\n"),
	}

	if _, err := applyGraphicsShim(directory, bundle); err != nil {
		t.Fatal(err)
	}

	paths := []string{
		filepath.Join(directory, dgvoodoo.D3D9_FILENAME),
		filepath.Join(directory, dgvoodoo.CONFIG_FILENAME),
	}

	// Use a known timestamp so an unnecessary rewrite is detectable.
	stamp := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, path := range paths {
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}

	got, err := applyGraphicsShim(directory, bundle)
	if err != nil {
		t.Fatal(err)
	}

	if got.State != GRAPHICS_SHIM_STATE_MANAGED {
		t.Fatalf("unexpected state: %s", got.State)
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}

		if !info.ModTime().Equal(stamp) {
			t.Errorf("%q was unexpectedly modified", path)
		}
	}

	assertFileContents(t, paths[0], bundle.D3D9)
	assertFileContents(t, paths[1], bundle.Config)
}
