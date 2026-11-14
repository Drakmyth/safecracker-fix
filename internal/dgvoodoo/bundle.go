package dgvoodoo

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"io"
)

const (
	VERSION           = "2.87.6"
	ARCHIVE_SHA256    = "9afa493396feec7108f42a2d97661eef96bbaac5addf0dfb5cfbc3001cd76a0a"
	D3D9_ARCHIVE_PATH = "MS/x86/D3D9.dll"
	D3D9_SHA256       = "060cb161eb7a0c608a520185f0b334db325ad71d69016bc26060c9d5b9651f7d"
	D3D9_FILENAME     = "D3D9.dll"
	CONFIG_FILENAME   = "dgVoodoo.conf"
)

//go:embed assets/dgVoodoo2_87_6.zip
var archiveData []byte

//go:embed assets/dgVoodoo.conf
var configData []byte

// Bundle contains the runtime files required for Safecracker's graphics compatibility fix.
type Bundle struct {
	D3D9   []byte
	Config []byte
}

// Load verifies the embedded upstream archive and extracts the graphics compatibility files.
func Load() (*Bundle, error) {
	return load(archiveData, configData)
}

func load(archive, config []byte) (*Bundle, error) {
	if actual := fmt.Sprintf("%x", sha256.Sum256(archive)); actual != ARCHIVE_SHA256 {
		return nil, fmt.Errorf("dgVoodoo archive hash mismatch: %s", actual)
	}

	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open dgVoodoo archive: %w", err)
	}

	dll, err := extractFile(reader, D3D9_ARCHIVE_PATH)
	if err != nil {
		return nil, err
	}

	if actual := fmt.Sprintf("%x", sha256.Sum256(dll)); actual != D3D9_SHA256 {
		return nil, fmt.Errorf("dgVoodoo D3D9.dll hash mismatch: %s", actual)
	}

	return &Bundle{D3D9: dll, Config: bytes.Clone(config)}, nil
}

// extractFile reads one exact archive entry without extracting any files to disk.
func extractFile(reader *zip.Reader, path string) ([]byte, error) {
	for _, entry := range reader.File {
		if entry.Name != path {
			continue
		}

		file, err := entry.Open()
		if err != nil {
			return nil, fmt.Errorf("open archive entry %q: %w", path, err)
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("read archive entry %q, %w", path, err)
		}

		return data, nil
	}

	return nil, fmt.Errorf("archive entry not found: %s", path)
}
