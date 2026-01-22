package ops

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestCreateAtomic(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Create File Success", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return &testutil.MockReadWriteCloser{}, nil
		}

		path, err := CreateAtomic(CreateOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Path:     "/test.txt",
			IsDir:    false,
			Conflict: ConflictOptions{Policy: conflict.Ask},
		})
		testutil.AssertNoError(t, err, "CreateAtomic should succeed")
		testutil.AssertEqual(t, "/test.txt", path, "Path should match")
	})

	t.Run("Create Directory Success", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		fs.MkdirAllFunc = func(ctx context.Context, path string, perm os.FileMode) error {
			return nil
		}

		path, err := CreateAtomic(CreateOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Path:     "/test_dir",
			IsDir:    true,
			Conflict: ConflictOptions{Policy: conflict.Ask},
		})
		testutil.AssertNoError(t, err, "CreateAtomic should succeed")
		testutil.AssertEqual(t, "/test_dir", path, "Path should match")
	})

	t.Run("Conflict Overwrite", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return &testutil.MockFileInfo{FName: "test.txt"}, nil
		}
		fs.RemoveAllFunc = func(ctx context.Context, path string) error {
			return nil
		}
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return &testutil.MockReadWriteCloser{}, nil
		}

		path, err := CreateAtomic(CreateOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Path:     "/test.txt",
			IsDir:    false,
			Conflict: ConflictOptions{Policy: conflict.Overwrite},
		})
		testutil.AssertNoError(t, err, "CreateAtomic should succeed")
		testutil.AssertEqual(t, "/test.txt", path, "Path should match")
	})
}
