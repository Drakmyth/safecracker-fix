package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceFile(t *testing.T) {
	original := []byte{0x00, 0x01, 0x02, 0x03}
	replacement := []byte{0xAA, 0xBB, 0xCC}

	tests := []struct {
		name        string
		original    []byte
		replacement []byte
		wantErr     bool
	}{
		{
			name:        "replace existing file",
			original:    original,
			replacement: replacement,
		},
		{
			name:        "replace with larger file",
			original:    original,
			replacement: []byte{1, 2, 3, 4, 5, 6},
		},
		{
			name:        "replace with empty file",
			original:    original,
			replacement: []byte{},
		},
		{
			name:        "replace empty file",
			original:    []byte{},
			replacement: replacement,
		},
		{
			name:        "replace with identical contents",
			original:    original,
			replacement: original,
		},
		{
			name:        "reject missing destination",
			replacement: replacement,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, EXECUTABLE_NAME)

			if tt.original != nil {
				if err := os.WriteFile(
					path,
					tt.original,
					0600,
				); err != nil {
					t.Fatal(err)
				}
			}

			err := replaceFile(path, tt.replacement)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v",
					err, tt.wantErr)
			}

			if tt.wantErr {
				assertFileContents(t, path, tt.original)
			} else {
				assertFileContents(t, path, tt.replacement)
			}

			assertNoTemporaryFiles(t, directory)
		})
	}
}

func TestReplaceFileRejectsDirectory(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, EXECUTABLE_NAME)

	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(path, []byte{0xAA}); err == nil {
		t.Fatal("expected error when destination is a directory")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if !info.IsDir() {
		t.Fatal("destination directory was modified")
	}

	assertNoTemporaryFiles(t, directory)
}

// assertNoTemporaryFiles verifies that replacement operations
// do not leave temporary files behind.
func assertNoTemporaryFiles(t *testing.T, directory string) {
	t.Helper()

	matches, err := filepath.Glob(
		filepath.Join(directory, ".safecracker-fix-*.tmp"),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 0 {
		t.Errorf("unexpected temporary files: %v", matches)
	}
}
