package ops

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestDeleteMultiple(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	// Mock successful deletion
	fs.RemoveAllFunc = func(ctx context.Context, path string) error {
		return nil
	}

	// Mock IsReadOnly
	fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
		return false, nil
	}

	paths := []string{"/test/a", "/test/b"}
	err := DeleteMultiple(DeleteOptions{
		OpCtx: OpContext{Context: ctx, FS: fs},
		Paths: paths,
		Trash: TrashOptions{UseTrash: false},
	})
	testutil.AssertNoError(t, err, "DeleteMultiple should succeed")

	t.Run("Progress and multiple items", func(t *testing.T) {
		progChan := make(chan core.Progress, 10)
		paths := []string{"/a", "/b", "/c"}
		err := DeleteMultiple(DeleteOptions{
			OpCtx: OpContext{Context: ctx, FS: fs, Progress: progChan},
			Paths: paths,
			Trash: TrashOptions{UseTrash: false},
		})
		testutil.AssertNoError(t, err, "Should succeed")

		// Drain progress
		count := 0
		for range progChan {
			count++
			if count >= 4 { // 3 items + final message
				break
			}
		}
	})

	t.Run("Empty paths", func(t *testing.T) {
		err := DeleteMultiple(DeleteOptions{
			OpCtx: OpContext{Context: ctx, FS: fs},
			Paths: []string{},
			Trash: TrashOptions{UseTrash: false},
		})
		testutil.AssertNoError(t, err, "Should not error on empty paths")
	})

	t.Run("Error in ValidateWritable", func(t *testing.T) {
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
			return true, nil
		}
		err := DeleteMultiple(DeleteOptions{
			OpCtx: OpContext{Context: ctx, FS: fs},
			Paths: []string{"/test/a"},
			Trash: TrashOptions{UseTrash: false},
		})
		if err == nil {
			t.Error("Expected error when filesystem is read-only")
		}
	})

	t.Run("Trash on remote", func(t *testing.T) {
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return false, nil }
		fs.IsLocalFunc = func() bool { return false }
		err := DeleteMultiple(DeleteOptions{
			OpCtx: OpContext{Context: ctx, FS: fs},
			Paths: []string{"/test/a"},
			Trash: TrashOptions{UseTrash: true},
		})
		if err == nil {
			t.Fatal("Expected error when trashing on remote")
		}
	})
}

func TestMoveMultiple_Conflict(t *testing.T) {
	ctx := context.Background()

	setup := func() *testutil.MockFileSystem {
		fs := testutil.NewMockFileSystem()
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
			return false, nil
		}
		return fs
	}

	sources := []string{"/src/a", "/src/b"}
	destDir := "/dest"

	t.Run("ConflictAsk returns error", func(t *testing.T) {
		fs := setup()
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{FName: "a"}, nil
			}
			return nil, os.ErrNotExist
		}
		err := MoveMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Ask, ApplyToAll: false},
		})
		if err == nil {
			t.Fatal("expected conflict error")
		}
		var target *conflict.ConflictError
		testutil.AssertErrorType(t, err, &target, "Should be ConflictError")
	})

	t.Run("ConflictSkip skips the conflicting file", func(t *testing.T) {
		fs := setup()
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{FName: "a"}, nil
			}
			if strings.HasPrefix(path, "/src/") {
				return &testutil.MockFileInfo{FName: "item"}, nil
			}
			return nil, os.ErrNotExist
		}
		// Mock MoveFunc to track calls
		moveCalled := false
		fs.RenameFunc = func(ctx context.Context, old, new string) error {
			if old == "/src/a" {
				moveCalled = true
			}
			return nil
		}

		err := MoveMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Skip, ApplyToAll: true},
		})
		testutil.AssertNoError(t, err, "Should succeed with skip")
		testutil.AssertEqual(t, false, moveCalled, "Conflicting file should NOT be moved")
	})
}

func TestCopyMultiple(t *testing.T) {
	ctx := context.Background()

	setup := func() *testutil.MockFileSystem {
		fs := testutil.NewMockFileSystem()
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
			return false, nil
		}
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if strings.HasPrefix(path, "/src/") {
				return &testutil.MockFileInfo{FName: "file", FIsDir: false}, nil
			}
			return nil, os.ErrNotExist
		}
		fs.LstatFunc = fs.StatFunc
		fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
			return testutil.NewMockFile("file", nil), nil
		}
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return testutil.NewMockFile("file", nil), nil
		}
		fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error {
			return nil
		}
		return fs
	}

	sources := []string{"/src/a", "/src/b"}
	destDir := "/dest"

	t.Run("Basic Success", func(t *testing.T) {
		fs := setup()
		err := CopyMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite, ApplyToAll: true},
		})
		testutil.AssertNoError(t, err, "CopyMultiple should succeed")
	})

	t.Run("Empty sources", func(t *testing.T) {
		fs := setup()
		err := CopyMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  []string{},
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Ask, ApplyToAll: false},
		})
		testutil.AssertNoError(t, err, "Should not error on empty sources")
	})

	t.Run("Conflict", func(t *testing.T) {
		fs := setup()
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{FName: "a"}, nil
			}
			if strings.HasPrefix(path, "/src/") {
				return &testutil.MockFileInfo{FName: "item"}, nil
			}
			return nil, os.ErrNotExist
		}
		err := CopyMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Ask, ApplyToAll: false},
		})
		if err == nil {
			t.Fatal("Expected conflict error")
		}
	})

	t.Run("ValidateWritable error", func(t *testing.T) {
		fs := setup()
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
			return true, nil
		}
		err := CopyMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite, ApplyToAll: false},
		})
		if err == nil {
			t.Error("Expected error when destination is read-only")
		}
	})

	t.Run("Context cancelled", func(t *testing.T) {
		fs := setup()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := CopyMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite, ApplyToAll: false},
		})
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})

	t.Run("Progress and Rename", func(t *testing.T) {
		fs := setup()
		progChan := make(chan core.Progress, 10)
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{FName: "a"}, nil
			}
			if strings.HasPrefix(path, "/src/") {
				return &testutil.MockFileInfo{FName: "item"}, nil
			}
			return nil, os.ErrNotExist
		}
		err := CopyMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs, Progress: progChan},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Rename, ApplyToAll: false},
		})
		testutil.AssertNoError(t, err, "Should succeed with Rename policy")

		close(progChan)
		for range progChan {
		}
	})

	t.Run("Copy error", func(t *testing.T) {
		fs := setup()
		fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
			return nil, fmt.Errorf("read error")
		}
		err := CopyMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite, ApplyToAll: false},
		})
		if err == nil {
			t.Error("Expected error from Copy")
		}
	})

	t.Run("ConflictSkip skips the conflicting file", func(t *testing.T) {
		fs := setup()
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{FName: "a"}, nil
			}
			if strings.HasPrefix(path, "/src/") {
				return &testutil.MockFileInfo{FName: "item"}, nil
			}
			return nil, os.ErrNotExist
		}
		copyCalled := false
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			if path == "/dest/a" {
				copyCalled = true
			}
			return testutil.NewMockFile("item", nil), nil
		}

		err := CopyMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Skip, ApplyToAll: true},
		})
		testutil.AssertNoError(t, err, "Should succeed with skip")
		testutil.AssertEqual(t, false, copyCalled, "Conflicting file should NOT be copied")
	})

	t.Run("Generic resolver error", func(t *testing.T) {
		fs := setup()
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return nil, fmt.Errorf("stat error")
			}
			if strings.HasPrefix(path, "/src/") {
				return &testutil.MockFileInfo{FName: "item"}, nil
			}
			return nil, os.ErrNotExist
		}
		err := CopyMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite, ApplyToAll: false},
		})
		if err == nil || !strings.Contains(err.Error(), "stat error") {
			t.Errorf("Expected stat error, got %v", err)
		}
	})
}

func TestMoveMultiple_Extra(t *testing.T) {
	ctx := context.Background()

	setup := func() *testutil.MockFileSystem {
		fs := testutil.NewMockFileSystem()
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return false, nil }
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if strings.HasPrefix(path, "/src/") {
				return &testutil.MockFileInfo{FName: "item"}, nil
			}
			return nil, os.ErrNotExist
		}
		fs.LstatFunc = fs.StatFunc
		fs.RenameFunc = func(ctx context.Context, old, new string) error { return nil }
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			return testutil.NewMockFile("item", nil), nil
		}
		fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
			return testutil.NewMockFile("item", nil), nil
		}
		fs.RemoveAllFunc = func(ctx context.Context, path string) error { return nil }
		return fs
	}

	sources := []string{"/src/a"}
	destDir := "/dest"

	t.Run("ValidateWritable error", func(t *testing.T) {
		fs := setup()
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return true, nil }
		err := MoveMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite, ApplyToAll: false},
		})
		if err == nil {
			t.Error("Expected error when destination is read-only")
		}
	})

	t.Run("Context cancelled", func(t *testing.T) {
		fs := setup()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := MoveMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite, ApplyToAll: false},
		})
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})

	t.Run("Progress and Rename", func(t *testing.T) {
		fs := setup()
		progChan := make(chan core.Progress, 10)
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{FName: "a"}, nil
			}
			if strings.HasPrefix(path, "/src/") {
				return &testutil.MockFileInfo{FName: "item"}, nil
			}
			return nil, os.ErrNotExist
		}
		err := MoveMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs, Progress: progChan},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Rename, ApplyToAll: false},
		})
		testutil.AssertNoError(t, err, "Should succeed with Rename policy")
	})

	t.Run("Move error", func(t *testing.T) {
		fs := setup()
		fs.RenameFunc = func(ctx context.Context, old, new string) error {
			return fmt.Errorf("move error")
		}
		err := MoveMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite, ApplyToAll: false},
		})
		if err == nil {
			t.Error("Expected error from Move")
		}
	})

	t.Run("Generic resolver error", func(t *testing.T) {
		fs := setup()
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return nil, fmt.Errorf("stat error")
			}
			if strings.HasPrefix(path, "/src/") {
				return &testutil.MockFileInfo{FName: "item"}, nil
			}
			return nil, os.ErrNotExist
		}
		err := MoveMultiple(BatchOptions{
			OpCtx:    OpContext{Context: ctx, FS: fs},
			SrcFS:    fs,
			Sources:  sources,
			DestDir:  destDir,
			Conflict: ConflictOptions{Policy: conflict.Overwrite, ApplyToAll: false},
		})
		if err == nil || !strings.Contains(err.Error(), "stat error") {
			t.Errorf("Expected stat error, got %v", err)
		}
	})
}

func TestCheckAndMarkProcessing(t *testing.T) {
	processing := make(map[string]bool)
	paths := []string{"/a", "/b"}

	ok := CheckAndMarkProcessing(processing, paths)
	testutil.AssertEqual(t, true, ok, "Should mark as processing")
	testutil.AssertEqual(t, true, processing["/a"], "/a should be processing")
	testutil.AssertEqual(t, true, processing["/b"], "/b should be processing")

	ok = CheckAndMarkProcessing(processing, []string{"/b", "/c"})
	testutil.AssertEqual(t, false, ok, "Should return false if any path is already processing")
	testutil.AssertEqual(t, true, processing["/b"], "/b should still be processing")
	testutil.AssertEqual(t, false, processing["/c"], "/c should NOT be marked if overall check failed")
}
