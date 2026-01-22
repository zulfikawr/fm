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
	fs.AbsFunc = func(path string) (string, error) { return path, nil }
	fs.BaseFunc = func(path string) string { return path }

	t.Run("Same source and destination", func(t *testing.T) {
		fs.AbsFunc = func(path string) (string, error) {
			return "/same/path", nil
		}
		err := CrossCopy(CopyOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Src:      "/same/path",
			Dst:      "/same/path",
			Conflict: ConflictOptions{Policy: conflict.Ask},
		})
		if err == nil {
			t.Fatal("expected error when src == dst")
		}
	})

	t.Run("Permission denied on Create", func(t *testing.T) {
		fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/src" {
				return &testutil.MockFileInfo{FName: "src", FIsDir: false}, nil
			}
			return nil, os.ErrNotExist
		}
		fs.StatFunc = fs.LstatFunc
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return nil, os.ErrPermission
		}
		err := CrossCopy(CopyOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Src:      "/src",
			Dst:      "/dst",
			Conflict: ConflictOptions{Policy: conflict.Ask},
		})
		if err == nil {
			t.Fatal("expected permission error")
		}
	})
}

func TestCrossCopy_Full(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.AbsFunc = func(path string) (string, error) { return path, nil }
	fs.BaseFunc = func(path string) string { return path }

	fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" {
			return &testutil.MockFileInfo{FName: "src", FIsDir: true}, nil
		}
		if path == "/src/file1.txt" {
			return &testutil.MockFileInfo{FName: "file1.txt", FIsDir: false, FSize: 10}, nil
		}
		return nil, os.ErrNotExist
	}
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" || path == "/src/file1.txt" {
			return fs.Lstat(ctx, path)
		}
		return nil, os.ErrNotExist
	}
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		if path == "/src" {
			return []os.FileInfo{&testutil.MockFileInfo{FName: "file1.txt", FIsDir: false, FSize: 10}}, nil
		}
		return []os.FileInfo{}, nil
	}
	fs.MkdirAllFunc = func(ctx context.Context, path string, perm os.FileMode) error { return nil }
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("file", nil), nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("file", make([]byte, 10)), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := CrossCopy(CopyOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		SrcFS:    fs,
		Src:      "/src",
		Dst:      "/dst",
		Conflict: ConflictOptions{Policy: conflict.Ask},
	})
	testutil.AssertNoError(t, err, "CrossCopy should succeed")
}

func TestCopy(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.AbsFunc = func(path string) (string, error) { return path, nil }
	fs.BaseFunc = func(path string) string { return path }

	fs.LstatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" {
			return &testutil.MockFileInfo{FName: "file1.txt", FIsDir: false, FSize: 10}, nil
		}
		return nil, os.ErrNotExist
	}
	fs.StatFunc = fs.LstatFunc
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("file", nil), nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("file", make([]byte, 10)), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := Copy(CopyOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		Src:      "/src",
		Dst:      "/dst",
		Conflict: ConflictOptions{Policy: conflict.Overwrite},
	})
	testutil.AssertNoError(t, err, "Copy should succeed")
}

func TestCopyFile(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.BaseFunc = func(path string) string { return path }

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" {
			return &testutil.MockFileInfo{FName: "file1.txt", FIsDir: false, FSize: 10}, nil
		}
		return nil, os.ErrNotExist
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("file", nil), nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("file", make([]byte, 10)), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := crossCopyFile(CopyOptions{
		OpCtx: OpContext{Context: ctx, FS: fs},
		SrcFS: fs,
		Src:   "/src",
		Dst:   "/dst",
	})
	testutil.AssertNoError(t, err, "crossCopyFile should succeed")
}

func TestCopyDir(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.AbsFunc = func(path string) (string, error) { return path, nil }
	fs.BaseFunc = func(path string) string { return path }

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/src" {
			return &testutil.MockFileInfo{FName: "src", FIsDir: true}, nil
		}
		if path == "/src/file1.txt" {
			return &testutil.MockFileInfo{FName: "file1.txt", FIsDir: false, FSize: 10}, nil
		}
		return nil, os.ErrNotExist
	}
	fs.ReadDirFunc = func(ctx context.Context, path string) ([]os.FileInfo, error) {
		if path == "/src" {
			return []os.FileInfo{&testutil.MockFileInfo{FName: "file1.txt", FIsDir: false, FSize: 10}}, nil
		}
		return []os.FileInfo{}, nil
	}
	fs.MkdirAllFunc = func(ctx context.Context, path string, perm os.FileMode) error { return nil }
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("file", nil), nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("file", make([]byte, 10)), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := crossCopyDir(CopyOptions{
		OpCtx: OpContext{Context: ctx, FS: fs},
		SrcFS: fs,
		Src:   "/src",
		Dst:   "/dst",
	})
	testutil.AssertNoError(t, err, "crossCopyDir should succeed")
}

func TestCrossCopy_Rename(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.AbsFunc = func(path string) (string, error) { return path, nil }
	fs.BaseFunc = func(path string) string { return path }

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/dst" {
			return &testutil.MockFileInfo{FName: "dst"}, nil
		}
		if path == "/dst (1)" {
			return nil, os.ErrNotExist
		}
		if path == "/src" {
			return &testutil.MockFileInfo{FName: "src", FSize: 10}, nil
		}
		return nil, os.ErrNotExist
	}
	fs.LstatFunc = fs.StatFunc
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("src", make([]byte, 10)), nil
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile("dst (1)", nil), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error { return nil }

	err := CrossCopy(CopyOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		SrcFS:    fs,
		Src:      "/src",
		Dst:      "/dst",
		Conflict: ConflictOptions{Policy: conflict.Rename},
	})
	testutil.AssertNoError(t, err, "CrossCopy should succeed with Rename")
}

func TestCrossCopy_Overwrite(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.AbsFunc = func(path string) (string, error) { return path, nil }
	fs.BaseFunc = func(path string) string { return path }

	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/dst" {
			return &testutil.MockFileInfo{FName: "dst"}, nil
		}
		if path == "/src" {
			return &testutil.MockFileInfo{FName: "src", FSize: 10}, nil
		}
		return nil, os.ErrNotExist
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

	err := CrossCopy(CopyOptions{
		OpCtx:    OpContext{Context: ctx, FS: fs},
		SrcFS:    fs,
		Src:      "/src",
		Dst:      "/dst",
		Conflict: ConflictOptions{Policy: conflict.Overwrite},
	})
	testutil.AssertNoError(t, err, "CrossCopy should succeed with Overwrite")
}
