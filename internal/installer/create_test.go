package installer

import (
	"path/filepath"
	"testing"
)

func TestWriteNewFileDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.bak")

	original := []byte{0x00, 0x01, 0x02}
	replacement := []byte{0xAA, 0xBB, 0xCC}

	if err := writeNewFile(path, original); err != nil {
		t.Fatalf("create backup: %v", err)
	}

	assertFileContents(t, path, original)

	if err := writeNewFile(path, replacement); err == nil {
		t.Fatal("expected error when backup already exists")
	}

	assertFileContents(t, path, original)
}
