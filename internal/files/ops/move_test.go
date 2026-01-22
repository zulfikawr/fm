package ops

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestMove(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Atomic Rename Success", func(t *testing.T) {
		fs.RenameFunc = func(ctx context.Context, old, new string) error {
			return nil
		}
		err := Move(CopyOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Src:      "/src",
			Dst:      "/dst",
			Conflict: ConflictOptions{Policy: conflict.Overwrite},
		})
		testutil.AssertNoError(t, err, "Move should succeed with Rename")
	})

	t.Run("Cross-device Fallback", func(t *testing.T) {
		fs.RenameFunc = func(ctx context.Context, old, new string) error {
			return os.ErrInvalid // Simulate cross-device rename failure
		}
		fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return &testutil.MockFileInfo{FName: "file", FIsDir: false, FSize: 10}, nil
		}
		fs.StatFunc = fs.LstatFunc
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return testutil.NewMockFile("file", nil), nil
		}
		fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
			return testutil.NewMockFile("file", make([]byte, 10)), nil
		}
		fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }
		fs.RemoveAllFunc = func(ctx context.Context, path string) error { return nil }

		err := Move(CopyOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Src:      "/src",
			Dst:      "/dst",
			Conflict: ConflictOptions{Policy: conflict.Overwrite},
		})
		testutil.AssertNoError(t, err, "Move should succeed with Copy+Delete fallback")
	})

	t.Run("Empty Path", func(t *testing.T) {
		err := Move(CopyOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			Src:      "",
			Dst:      "/dst",
			Conflict: ConflictOptions{Policy: conflict.Ask},
		})
		if err == nil {
			t.Error("Expected error for empty source")
		}
	})
}

func TestCrossMove_RenameError(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.RenameFunc = func(ctx context.Context, old, new string) error {
		return os.ErrPermission
	}
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{FName: "file", FSize: 10}, nil
	}
	fs.LstatFunc = fs.StatFunc
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("file", make([]byte, 10)), nil
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("file", nil), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }
	fs.RemoveAllFunc = func(ctx context.Context, path string) error { return nil }
	fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return false, nil }

	err := CrossMove(CopyOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		SrcFS:    fs,
		Src:      "/src",
		Dst:      "/dst",
		Conflict: ConflictOptions{Policy: conflict.Overwrite},
	})
	testutil.AssertNoError(t, err, "Should succeed with fallback")
}

func TestCrossMove_Conflict(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return false, nil }
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/dst" {
			return &testutil.MockFileInfo{FName: "dst"}, nil
		}
		return nil, os.ErrNotExist
	}

	err := CrossMove(CopyOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		SrcFS:    fs,
		Src:      "/src",
		Dst:      "/dst",
		Conflict: ConflictOptions{Policy: conflict.Ask},
	})
	if err == nil {
		t.Fatal("Expected conflict error")
	}

	t.Run("ConflictSkip", func(t *testing.T) {
		err := CrossMove(CopyOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Src:      "/src",
			Dst:      "/dst",
			Conflict: ConflictOptions{Policy: conflict.Skip},
		})
		testutil.AssertNoError(t, err, "Should succeed with skip")
	})

	t.Run("Copy Error Cleanup", func(t *testing.T) {
		calledRemove := false
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/src" {
				return &testutil.MockFileInfo{FName: "src", FIsDir: false, FSize: 10}, nil
			}
			return nil, os.ErrNotExist
		}
		fs.LstatFunc = fs.StatFunc
		fs.RenameFunc = func(ctx context.Context, old, new string) error { return os.ErrInvalid }
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return nil, os.ErrPermission
		}
		fs.RemoveAllFunc = func(ctx context.Context, path string) error {
			calledRemove = true
			return nil
		}

		err := CrossMove(CopyOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Src:      "/src",
			Dst:      "/dst",
			Conflict: ConflictOptions{Policy: conflict.Overwrite},
		})
		if err == nil {
			t.Fatal("Expected copy error")
		}
		if !calledRemove {
			t.Error("fs.RemoveAll was not called for cleanup")
		}
	})
}

func TestVerifyCrossMove(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Size Match", func(t *testing.T) {
		fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return &testutil.MockFileInfo{FSize: 100, FIsDir: false}, nil
		}
		err := verifyCrossMove(CopyOptions{
			OpCtx: OpContext{Context: ctx, FS: fs},
			SrcFS: fs,
			Src:   "/src",
			Dst:   "/dst",
		})
		testutil.AssertNoError(t, err, "Should not error when sizes match")
	})

	t.Run("Size Mismatch", func(t *testing.T) {
		fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/src" {
				return &testutil.MockFileInfo{FSize: 100, FIsDir: false}, nil
			}
			return &testutil.MockFileInfo{FSize: 200, FIsDir: false}, nil
		}
		err := verifyCrossMove(CopyOptions{
			OpCtx: OpContext{Context: ctx, FS: fs},
			SrcFS: fs,
			Src:   "/src",
			Dst:   "/dst",
		})
		if err == nil {
			t.Fatal("Expected error for size mismatch")
		}
	})
}
