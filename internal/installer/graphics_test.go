package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Drakmyth/safecracker-fix/internal/dgvoodoo"
)

func TestInspectGraphics(t *testing.T) {
	bundle := &dgvoodoo.Bundle{
		D3D9:   []byte{0x01, 0x02, 0x03},
		Config: []byte(dgvoodoo.CONFIG_OWNERSHIP_MARKER + "\n\nVersion = 0x287\n"),
	}

	otherDLL := []byte{0xAA, 0xBB, 0xCC}
	otherConfig := []byte("Version = 0x287\n")

	modifiedConfig := append(bytes.Clone(bundle.Config), []byte("\n[DirectX]\nFastVideoMemoryAccess = true\n")...)

	tests := []struct {
		name   string
		dll    []byte
		config []byte
		want   GraphicsInspection
	}{
		{
			name: "not installed",
			want: GraphicsInspection{
				State: GRAPHICS_STATE_NOT_INSTALLED,
			},
		},
		{
			name:   "managed installation",
			dll:    bundle.D3D9,
			config: bundle.Config,
			want: GraphicsInspection{
				State:         GRAPHICS_STATE_MANAGED,
				DLLExists:     true,
				DLLMatches:    true,
				ConfigExists:  true,
				ConfigManaged: true,
				ConfigMatches: true,
			},
		},
		{
			name:   "managed installation with modified config",
			dll:    bundle.D3D9,
			config: modifiedConfig,
			want: GraphicsInspection{
				State:         GRAPHICS_STATE_MANAGED,
				DLLExists:     true,
				DLLMatches:    true,
				ConfigExists:  true,
				ConfigManaged: true,
			},
		},
		{
			name: "external installation without config",
			dll:  bundle.D3D9,
			want: GraphicsInspection{
				State:      GRAPHICS_STATE_EXTERNAL,
				DLLExists:  true,
				DLLMatches: true,
			},
		},
		{
			name:   "external installation with config",
			dll:    bundle.D3D9,
			config: otherConfig,
			want: GraphicsInspection{
				State:        GRAPHICS_STATE_EXTERNAL,
				DLLExists:    true,
				DLLMatches:   true,
				ConfigExists: true,
			},
		},
		{
			name:   "incomplete managed installation",
			config: bundle.Config,
			want: GraphicsInspection{
				State:         GRAPHICS_STATE_INCOMPLETE,
				ConfigExists:  true,
				ConfigManaged: true,
				ConfigMatches: true,
			},
		},
		{
			name:   "incomplete with modified managed config",
			config: modifiedConfig,
			want: GraphicsInspection{
				State:         GRAPHICS_STATE_INCOMPLETE,
				ConfigExists:  true,
				ConfigManaged: true,
			},
		},
		{
			name:   "conflicting config without DLL",
			config: otherConfig,
			want: GraphicsInspection{
				State:        GRAPHICS_STATE_CONFLICT,
				ConfigExists: true,
			},
		},
		{
			name: "conflicting DLL without config",
			dll:  otherDLL,
			want: GraphicsInspection{
				State:     GRAPHICS_STATE_CONFLICT,
				DLLExists: true,
			},
		},
		{
			name:   "conflicting DLL with external config",
			dll:    otherDLL,
			config: otherConfig,
			want: GraphicsInspection{
				State:        GRAPHICS_STATE_CONFLICT,
				DLLExists:    true,
				ConfigExists: true,
			},
		},
		{
			name:   "conflicting DLL with managed config",
			dll:    otherDLL,
			config: bundle.Config,
			want: GraphicsInspection{
				State:         GRAPHICS_STATE_CONFLICT,
				DLLExists:     true,
				ConfigExists:  true,
				ConfigManaged: true,
				ConfigMatches: true,
			},
		},
		{
			name: "empty DLL",
			dll:  []byte{},
			want: GraphicsInspection{
				State:     GRAPHICS_STATE_CONFLICT,
				DLLExists: true,
			},
		},
		{
			name:   "empty config",
			config: []byte{},
			want: GraphicsInspection{
				State:        GRAPHICS_STATE_CONFLICT,
				ConfigExists: true,
			},
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

			got, err := inspectGraphics(directory, bundle)
			if err != nil {
				t.Fatal(err)
			}

			if got == nil {
				t.Fatal("expected graphics inspection")
			}

			if *got != tt.want {
				t.Errorf("inspection = %+v, want %+v", *got, tt.want)
			}

			// Inspection must not modify either file.
			assertFileContents(t, dllPath, tt.dll)
			assertFileContents(t, configPath, tt.config)
		})
	}
}

func TestInspectGraphicsReadError(t *testing.T) {
	bundle := &dgvoodoo.Bundle{
		D3D9:   []byte{0x01, 0x02, 0x03},
		Config: []byte(dgvoodoo.CONFIG_OWNERSHIP_MARKER),
	}

	for _, filename := range []string{
		dgvoodoo.D3D9_FILENAME,
		dgvoodoo.CONFIG_FILENAME,
	} {
		t.Run(filename, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, filename)

			if err := os.Mkdir(path, 0700); err != nil {
				t.Fatal(err)
			}

			got, err := inspectGraphics(directory, bundle)

			if err == nil {
				t.Fatal("expected read error")
			}

			if got != nil {
				t.Fatal("expected nil inspection on error")
			}
		})
	}
}
