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
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("failed to remove temp dir %s: %v", dir, err)
		}
	})
	return dir
}
