package ops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files/local"
	"fm/internal/testutil"
)

func TestDelete(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	fs := &local.LocalFS{}

	t.Run("Delete File", func(t *testing.T) {
		path := filepath.Join(tmpDir, "todelete.txt")
		testutil.CreateTestFile(t, tmpDir, "todelete.txt", "delete me")
		if err := Delete(context.Background(), fs, path, nil); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		if _, err := os.Stat(path); err == nil {
			t.Error("File still exists after delete")
		}
	})
}

func TestTrash(t *testing.T) {
	// Trash is hard to test cross-platform in CI without side effects
	// but we can test the error cases.
	m := testutil.NewMockFileSystem()
	m.IsLocalFunc = func() bool { return false }

	err := Trash(context.Background(), m, "/remote/path")
	if err == nil {
		t.Error("Expected error when trashing on remote filesystem")
	}
}
