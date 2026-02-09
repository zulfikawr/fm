package ops

import (
	"context"
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestDelete(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Successful Delete", func(t *testing.T) {
		called := false
		fs.RemoveAllFunc = func(ctx context.Context, path string) error {
			called = true
			return nil
		}

		progChan := make(chan core.Progress, 2)
		err := Delete(DeleteOptions{
			OpCtx: OpContext{
				Context:  ctx,
				FS:       fs,
				Progress: progChan,
			},
			Paths: []string{"/path/to/delete"},
		})

		testutil.AssertNoError(t, err, "Delete should succeed")
		if !called {
			t.Error("fs.RemoveAll was not called")
		}

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
		err := Delete(DeleteOptions{
			OpCtx: OpContext{Context: ctx, FS: fs},
			Paths: []string{""},
		})
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})
}

func TestMoveToTrash(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Remote FS Unsupported", func(t *testing.T) {
		fs.IsLocalFunc = func() bool {
			return false
		}
		err := MoveToTrash(ctx, fs, "/path/to/trash")
		if err == nil {
			t.Fatal("expected error for non-local filesystem")
		}
	})

	t.Run("Empty Path", func(t *testing.T) {
		err := MoveToTrash(ctx, fs, "")
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})
}
