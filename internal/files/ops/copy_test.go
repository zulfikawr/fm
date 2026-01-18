package ops

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestCrossCopy_EdgeCases(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Same source and destination", func(t *testing.T) {
		fs.AbsFunc = func(path string) (string, error) {
			return "/same/path", nil
		}
		err := CrossCopy(ctx, fs, fs, "/same/path", "/same/path", nil, conflict.Ask)
		if err == nil {
			t.Fatal("expected error when src == dst")
		}
	})

	t.Run("Permission denied on Create", func(t *testing.T) {
		fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return &testutil.MockFileInfo{NameStr: "src", IsDirBool: false}, nil
		}
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return nil, os.ErrPermission
		}
		err := CrossCopy(ctx, fs, fs, "/src", "/dst", nil, conflict.Ask)
		if err == nil {
			t.Fatal("expected permission error")
		}
	})
}

func TestCrossCopy_Full(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	// Setup for recursive copy test
	fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" {
			return &testutil.MockFileInfo{NameStr: "src", IsDirBool: true}, nil
		}
		if path == "/src/file1.txt" {
			return &testutil.MockFileInfo{NameStr: "file1.txt", IsDirBool: false, SizeInt: 10}, nil
		}
		return nil, os.ErrNotExist
	}
	fs.StatFunc = fs.LstatFunc
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		if path == "/src" {
			return []os.FileInfo{&testutil.MockFileInfo{NameStr: "file1.txt", IsDirBool: false, SizeInt: 10}}, nil
		}
		return []os.FileInfo{}, nil
	}
	fs.MkdirAllFunc = func(ctx context.Context, path string, perm os.FileMode) error { return nil }
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile(fs.Base(path), nil), nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile(fs.Base(path), make([]byte, 10)), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := CrossCopy(ctx, fs, fs, "/src", "/dst", nil, conflict.Ask)
	testutil.AssertNoError(t, err, "CrossCopy should succeed")

	fs.AssertCalled(t, "MkdirAll")
	fs.AssertCalled(t, "Create")
}

func TestCopy(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{NameStr: "file1.txt", IsDirBool: false, SizeInt: 10}, nil
	}
	fs.StatFunc = fs.LstatFunc
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile(fs.Base(path), nil), nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile(fs.Base(path), make([]byte, 10)), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := Copy(ctx, fs, "/src", "/dst", nil, conflict.Overwrite)
	testutil.AssertNoError(t, err, "Copy should succeed")
}

func TestCopyFile(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{NameStr: "file1.txt", IsDirBool: false, SizeInt: 10}, nil
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile(fs.Base(path), nil), nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile(fs.Base(path), make([]byte, 10)), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := copyFile(ctx, fs, "/src", "/dst", nil)
	testutil.AssertNoError(t, err, "copyFile should succeed")
}

func TestCopyDir(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" {
			return &testutil.MockFileInfo{NameStr: "src", IsDirBool: true}, nil
		}
		return &testutil.MockFileInfo{NameStr: "file1.txt", IsDirBool: false, SizeInt: 10}, nil
	}
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		if path == "/src" {
			return []os.FileInfo{&testutil.MockFileInfo{NameStr: "file1.txt", IsDirBool: false, SizeInt: 10}}, nil
		}
		return []os.FileInfo{}, nil
	}
	fs.MkdirAllFunc = func(ctx context.Context, path string, perm os.FileMode) error { return nil }
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile(fs.Base(path), nil), nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile(fs.Base(path), make([]byte, 10)), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := copyDir(ctx, fs, "/src", "/dst", nil)
	testutil.AssertNoError(t, err, "copyDir should succeed")
}

func TestCrossCopy_Rename(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/dst" {
			return &testutil.MockFileInfo{NameStr: "dst"}, nil
		}
		if path == "/dst (1)" {
			return nil, os.ErrNotExist
		}
		return &testutil.MockFileInfo{NameStr: "src", SizeInt: 10}, nil
	}
	fs.LstatFunc = fs.StatFunc
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("src", make([]byte, 10)), nil
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("dst (1)", nil), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := CrossCopy(ctx, fs, fs, "/src", "/dst", nil, conflict.Rename)
	testutil.AssertNoError(t, err, "CrossCopy should succeed with Rename")
}

func TestCrossCopy_Overwrite(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/dst" {
			return &testutil.MockFileInfo{NameStr: "dst"}, nil
		}
		return &testutil.MockFileInfo{NameStr: "src", SizeInt: 10}, nil
	}
	fs.LstatFunc = fs.StatFunc
	fs.RemoveAllFunc = func(ctx context.Context, path string) error { return nil }
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("src", make([]byte, 10)), nil
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("dst", nil), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := CrossCopy(ctx, fs, fs, "/src", "/dst", nil, conflict.Overwrite)
	testutil.AssertNoError(t, err, "CrossCopy should succeed with Overwrite")
	fs.AssertCalled(t, "RemoveAll")
}
