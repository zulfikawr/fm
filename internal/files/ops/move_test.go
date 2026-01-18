package ops

import (
	"context"
	"io"
	"os"
	"testing"

	"fm/internal/files/conflict"
	"fm/internal/testutil"
)

func TestMove(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Atomic Rename Success", func(t *testing.T) {
		fs.RenameFunc = func(ctx context.Context, old, new string) error {
			return nil
		}
		err := Move(ctx, fs, "/src", "/dst", nil, conflict.Overwrite)
		testutil.AssertNoError(t, err, "Move should succeed with Rename")
	})

	t.Run("Cross-device Fallback", func(t *testing.T) {
		fs.RenameFunc = func(ctx context.Context, old, new string) error {
			return os.ErrInvalid // Simulate cross-device rename failure
		}
		fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return &testutil.MockFileInfo{NameStr: fs.Base(path), IsDirBool: false, SizeInt: 10}, nil
		}
		fs.StatFunc = fs.LstatFunc
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return testutil.NewMockFile(fs.Base(path), nil), nil
		}
		fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
			return testutil.NewMockFile(fs.Base(path), make([]byte, 10)), nil
		}
		fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }
		fs.RemoveAllFunc = func(ctx context.Context, path string) error { return nil }

		err := Move(ctx, fs, "/src", "/dst", nil, conflict.Overwrite)
		testutil.AssertNoError(t, err, "Move should succeed with Copy+Delete fallback")
	})

	t.Run("Empty Path", func(t *testing.T) {
		err := Move(ctx, fs, "", "/dst", nil, conflict.Ask)
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
		return &testutil.MockFileInfo{NameStr: "file", SizeInt: 10}, nil
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

	err := CrossMove(ctx, fs, fs, "/src", "/dst", nil, conflict.Overwrite)
	testutil.AssertNoError(t, err, "Should succeed with fallback")
}

func TestCrossMove_Conflict(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return false, nil }
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/dst" {
			return &testutil.MockFileInfo{NameStr: "dst"}, nil
		}
		return nil, os.ErrNotExist
	}

	err := CrossMove(ctx, fs, fs, "/src", "/dst", nil, conflict.Ask)
	if err == nil {
		t.Fatal("Expected conflict error")
	}

	t.Run("ConflictSkip", func(t *testing.T) {
		err := CrossMove(ctx, fs, fs, "/src", "/dst", nil, conflict.Skip)
		testutil.AssertNoError(t, err, "Should succeed with skip")
	})

	t.Run("Copy Error Cleanup", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/src" {
				return &testutil.MockFileInfo{NameStr: "src", IsDirBool: false, SizeInt: 10}, nil
			}
			return nil, os.ErrNotExist
		}
		fs.LstatFunc = fs.StatFunc
		fs.RenameFunc = func(ctx context.Context, old, new string) error { return os.ErrInvalid }
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return nil, os.ErrPermission
		}
		fs.RemoveAllFunc = func(ctx context.Context, path string) error { return nil }

		err := CrossMove(ctx, fs, fs, "/src", "/dst", nil, conflict.Overwrite)
		if err == nil {
			t.Fatal("Expected copy error")
		}
		fs.AssertCalled(t, "RemoveAll")
	})
}

func TestVerifyCrossMove(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Size Match", func(t *testing.T) {
		fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return &testutil.MockFileInfo{SizeInt: 100, IsDirBool: false}, nil
		}
		err := verifyCrossMove(ctx, fs, fs, "/src", "/dst")
		testutil.AssertNoError(t, err, "Should not error when sizes match")
	})

	t.Run("Size Mismatch", func(t *testing.T) {
		fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/src" {
				return &testutil.MockFileInfo{SizeInt: 100, IsDirBool: false}, nil
			}
			return &testutil.MockFileInfo{SizeInt: 200, IsDirBool: false}, nil
		}
		err := verifyCrossMove(ctx, fs, fs, "/src", "/dst")
		if err == nil {
			t.Fatal("Expected error for size mismatch")
		}
	})
}

func TestVerifyMove_Deprecated(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{SizeInt: 100, IsDirBool: false}, nil
	}
	err := verifyMove(ctx, fs, "/src", "/dst")
	testutil.AssertNoError(t, err, "Should succeed")
}
