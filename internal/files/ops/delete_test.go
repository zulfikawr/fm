package ops

import (
	"context"
	"testing"

	"fm/internal/files/core"
	"fm/internal/testutil"
)

func TestDelete(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Successful Delete", func(t *testing.T) {
		fs.RemoveAllFunc = func(ctx context.Context, path string) error {
			return nil
		}

		progChan := make(chan core.Progress, 2)
		err := Delete(ctx, fs, "/path/to/delete", progChan)

		testutil.AssertNoError(t, err, "Delete should succeed")
		fs.AssertCalled(t, "RemoveAll")

		// Check progress messages
		p1 := <-progChan
		if p1.Percent != 0 {
			t.Errorf("Expected 0%% progress, got %f", p1.Percent)
		}
		p2 := <-progChan
		if p2.Percent != 1.0 {
			t.Errorf("Expected 100%% progress, got %f", p2.Percent)
		}
	})

	t.Run("Empty Path", func(t *testing.T) {
		err := Delete(ctx, fs, "", nil)
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})
}

func TestTrash(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Remote FS Unsupported", func(t *testing.T) {
		fs.IsLocalFunc = func() bool {
			return false
		}
		err := Trash(ctx, fs, "/path/to/trash")
		if err == nil {
			t.Fatal("expected error for non-local filesystem")
		}
	})

	t.Run("Empty Path", func(t *testing.T) {
		err := Trash(ctx, fs, "")
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})
}
