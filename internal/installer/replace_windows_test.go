package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceFileLockedDestination(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, EXECUTABLE_NAME)

	original := []byte{0x00, 0x01, 0x02, 0x03}
	replacement := []byte{0xAA, 0xBB, 0xCC, 0xDD}

	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}

	// On Windows, os.Open does not grant delete sharing.
	// Keeping this handle open prevents replacement.
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if err := replaceFile(path, replacement); err == nil {
		t.Fatal("expected error when destination is locked")
	}

	assertFileContents(t, path, original)
	assertNoTemporaryFiles(t, directory)

	// Once the handle is closed, replacement should succeed.
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(path, replacement); err != nil {
		t.Fatalf("replace after releasing lock: %v", err)
	}

	assertFileContents(t, path, replacement)
	assertNoTemporaryFiles(t, directory)
}
