package testutil

import (
	"os"
	"path/filepath"
)

// TempFolder manages a temporary directory for testing.
type TempFolder struct {
	Path string
	t    TB
}

// NewTempFolder creates a new temporary directory and returns a TempFolder.
func NewTempFolder(t TB) *TempFolder {
	t.Helper()
	dir, err := os.MkdirTemp("", "fm-test-*")
	if err != nil {
		t.Fatalf("failed to create temp folder: %v", err)
	}
	return &TempFolder{
		Path: dir,
		t:    t,
	}
}

// Cleanup removes the temporary directory.
func (f *TempFolder) Cleanup() {
	f.t.Helper()
	os.RemoveAll(f.Path)
}

// Join returns a path within the temporary folder.
func (f *TempFolder) Join(elem ...string) string {
	f.t.Helper()
	fullPath := []string{f.Path}
	fullPath = append(fullPath, elem...)
	return filepath.Join(fullPath...)
}

// WriteFile creates a file with the given content in the temporary folder.
func (f *TempFolder) WriteFile(name, content string) string {
	f.t.Helper()
	path := f.Join(name)
	// Ensure parent directory exists
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		f.t.Fatalf("failed to create directory for file: %v", err)
	}
	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		f.t.Fatalf("failed to write test file: %v", err)
	}
	return path
}

// Mkdir creates a directory within the temporary folder.
func (f *TempFolder) Mkdir(name string) string {
	f.t.Helper()
	path := f.Join(name)
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		f.t.Fatalf("failed to create test directory: %v", err)
	}
	return path
}
