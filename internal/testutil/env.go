package testutil

import (
	"os"
	"testing"
)

// TempDir creates a temporary directory and registers a cleanup function.
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "fm-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return dir
}
