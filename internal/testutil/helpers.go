package testutil

import (
	"errors"
	"os"
	"path/filepath"
)

// TB is a subset of testing.TB that we use for helpers to be compatible with rapid.T
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
	Errorf(format string, args ...any)
}

// TempDir creates a temporary directory and returns its path and a cleanup function
func TempDir(t TB) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "fm-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	return dir, func() {
		os.RemoveAll(dir)
	}
}

// CreateTestFile creates a file with the given content in the specified directory
func CreateTestFile(t TB, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	return path
}

// AssertErrorType checks if an error is of a specific type
func AssertErrorType(t TB, err error, target interface{}, msg string) {
	t.Helper()
	if !errors.As(err, target) {
		t.Errorf("%s: expected error type %T, got %T (%v)", msg, target, err, err)
	}
}
