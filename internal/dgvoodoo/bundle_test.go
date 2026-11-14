package dgvoodoo

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestLoad(t *testing.T) {
	bundle, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if bundle == nil {
		t.Fatal("expected bundle")
	}

	if actual := fmt.Sprintf("%x", sha256.Sum256(bundle.D3D9)); actual != D3D9_SHA256 {
		t.Errorf("D3D9 hash = %s, want %s", actual, D3D9_SHA256)
	}

	if !bytes.Equal(bundle.Config, configData) {
		t.Error("configuration does not match embedded asset")
	}

	if len(bundle.Config) == 0 {
		t.Error("configuration is empty")
	}
}

func TestLoadRejectsInvalidArchive(t *testing.T) {
	tests := []struct {
		name    string
		archive []byte
	}{
		{
			name: "nil archive",
		},
		{
			name:    "empty archive",
			archive: []byte{},
		},
		{
			name:    "invalid archive",
			archive: []byte{0x00, 0x01, 0x02},
		},
		{
			name:    "truncated archive",
			archive: archiveData[:len(archiveData)/2],
		},
		{
			name: "modified archive",
			archive: func() []byte {
				data := bytes.Clone(archiveData)
				data[0] ^= 0xFF
				return data
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle, err := load(tt.archive, configData)

			if err == nil {
				t.Fatal("expected error")
			}

			if bundle != nil {
				t.Fatal("expected nil bundle on error")
			}
		})
	}
}

func TestLoadReturnsIndependentCopies(t *testing.T) {
	first, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	second, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	expectedDLL := bytes.Clone(second.D3D9)
	expectedConfig := bytes.Clone(second.Config)

	// Modify the first bundle.
	first.D3D9[0] ^= 0xFF
	first.Config[0] ^= 0xFF

	// Existing bundles must remain unchanged.
	if !bytes.Equal(second.D3D9, expectedDLL) {
		t.Error("DLL data is shared between bundles")
	}

	if !bytes.Equal(second.Config, expectedConfig) {
		t.Error("configuration is shared between bundles")
	}

	// Future loads must also remain unaffected.
	third, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(third.D3D9, expectedDLL) {
		t.Error("embedded DLL was modified")
	}

	if !bytes.Equal(third.Config, expectedConfig) {
		t.Error("embedded configuration was modified")
	}
}

func TestExtractFile(t *testing.T) {
	const (
		X86_PATH = "MS/x86/D3D9.dll"
		X64_PATH = "MS/x64/D3D9.dll"
	)

	x86Data := []byte{0x01, 0x02, 0x03}
	x64Data := []byte{0xAA, 0xBB, 0xCC}

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)

	entries := map[string][]byte{
		X86_PATH: x86Data,
		X64_PATH: x64Data,
	}

	for path, data := range entries {
		file, err := writer.Create(path)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := file.Write(data); err != nil {
			t.Fatal(err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	reader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		want    []byte
		wantErr bool
	}{
		{
			name: "extract x86 DLL",
			path: X86_PATH,
			want: x86Data,
		},
		{
			name: "extract x64 DLL",
			path: X64_PATH,
			want: x64Data,
		},
		{
			name:    "reject missing entry",
			path:    "missing.dll",
			wantErr: true,
		},
		{
			name:    "require exact path",
			path:    "D3D9.dll",
			wantErr: true,
		},
		{
			name:    "reject incorrect case",
			path:    "MS/x86/d3d9.dll",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := extractFile(reader, tt.path)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if data != nil {
					t.Fatal("expected nil data on error")
				}
				return
			}

			if !bytes.Equal(data, tt.want) {
				t.Errorf("data = %x, want %x", data, tt.want)
			}
		})
	}
}

func TestIsManagedConfig(t *testing.T) {
	tests := []struct {
		name string
		data string
		want bool
	}{
		{
			name: "empty configuration",
		},
		{
			name: "unmarked configuration",
			data: "Version = 0x287\n",
		},
		{
			name: "exact marker",
			data: CONFIG_OWNERSHIP_MARKER,
			want: true,
		},
		{
			name: "embedded configuration",
			data: string(configData),
			want: true,
		},
		{
			name: "Windows line endings",
			data: CONFIG_OWNERSHIP_MARKER + "\r\nVersion = 0x287\r\n",
			want: true,
		},
		{
			name: "surrounding whitespace",
			data: "  " + CONFIG_OWNERSHIP_MARKER + "  \n",
			want: true,
		},
		{
			name: "marker after other settings",
			data: "Version = 0x287\n" + CONFIG_OWNERSHIP_MARKER,
			want: true,
		},
		{
			name: "partial marker",
			data: "; Managed by Safecracker",
		},
		{
			name: "marker with additional text",
			data: CONFIG_OWNERSHIP_MARKER + " (old)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsManagedConfig([]byte(tt.data)); got != tt.want {
				t.Errorf("IsManagedConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
