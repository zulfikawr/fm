package ops

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"fm/internal/files/conflict"
	"fm/internal/files/core"
	"fm/internal/testutil"
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
	err := DeleteMultiple(ctx, fs, paths, false, nil)
	testutil.AssertNoError(t, err, "DeleteMultiple should succeed")

	t.Run("Progress and multiple items", func(t *testing.T) {
		progChan := make(chan core.Progress, 10)
		paths := []string{"/a", "/b", "/c"}
		err := DeleteMultiple(ctx, fs, paths, false, progChan)
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
		err := DeleteMultiple(ctx, fs, []string{}, false, nil)
		testutil.AssertNoError(t, err, "Should not error on empty paths")
	})

	t.Run("Error in ValidateWritable", func(t *testing.T) {
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
			return true, nil
		}
		err := DeleteMultiple(ctx, fs, []string{"/test/a"}, false, nil)
		if err == nil {
			t.Error("Expected error when filesystem is read-only")
		}
	})

	t.Run("Trash on remote", func(t *testing.T) {
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return false, nil }
		fs.IsLocalFunc = func() bool { return false }
		err := DeleteMultiple(ctx, fs, []string{"/test/a"}, true, nil)
		if err == nil {
			t.Fatal("Expected error when trashing on remote")
		}
	})
}

func TestMoveMultiple_Conflict(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	// Mock destination existence
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		if path == "/dest/a" {
			return &testutil.MockFileInfo{NameStr: "a"}, nil
		}
		return nil, os.ErrNotExist
	}

	fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
		return false, nil
	}

	sources := []string{"/src/a", "/src/b"}
	destDir := "/dest"

	t.Run("ConflictAsk returns error", func(t *testing.T) {
		err := MoveMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Ask)
		if err == nil {
			t.Fatal("expected conflict error")
		}
		var target *conflict.ConflictError
		testutil.AssertErrorType(t, err, &target, "Should be ConflictError")
	})

	t.Run("ConflictSkip skips the conflicting file", func(t *testing.T) {
		// Mock MoveFunc to track calls
		moveCalled := false
		fs.RenameFunc = func(ctx context.Context, old, new string) error {
			if old == "/src/a" {
				moveCalled = true
			}
			return nil
		}

		err := MoveMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Skip)
		testutil.AssertNoError(t, err, "Should succeed with skip")
		testutil.AssertEqual(t, false, moveCalled, "Conflicting file should NOT be moved")
	})
}

func TestCopyMultiple(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
		return false, nil
	}
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{NameStr: fs.Base(path), IsDirBool: false}, nil
	}
	fs.LstatFunc = fs.StatFunc
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile(fs.Base(path), nil), nil
	}
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile(fs.Base(path), nil), nil
	}
	fs.ChmodFunc = func(ctx context.Context, path string, mode os.FileMode) error {
		return nil
	}

	sources := []string{"/src/a", "/src/b"}
	destDir := "/dest"

	err := CopyMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Overwrite)
	testutil.AssertNoError(t, err, "CopyMultiple should succeed")

	t.Run("Empty sources", func(t *testing.T) {
		err := CopyMultiple(ctx, fs, fs, []string{}, destDir, nil, conflict.Ask)
		testutil.AssertNoError(t, err, "Should not error on empty sources")
	})

	t.Run("Conflict", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{NameStr: "a"}, nil
			}
			return nil, os.ErrNotExist
		}
		err := CopyMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Ask)
		if err == nil {
			t.Fatal("Expected conflict error")
		}
	})

	t.Run("ValidateWritable error", func(t *testing.T) {
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
			return true, nil
		}
		err := CopyMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Overwrite)
		if err == nil {
			t.Error("Expected error when destination is read-only")
		}
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return false, nil }
	})

	t.Run("Context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := CopyMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Overwrite)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})

	t.Run("Progress and Rename", func(t *testing.T) {
		progChan := make(chan core.Progress, 10)
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{NameStr: "a"}, nil
			}
			if path == "/src/a" || path == "/src/b" {
				return &testutil.MockFileInfo{NameStr: fs.Base(path)}, nil
			}
			return nil, os.ErrNotExist
		}
		// Policy Rename will cause isRenamed to be true
		err := CopyMultiple(ctx, fs, fs, sources, destDir, progChan, conflict.Rename)
		testutil.AssertNoError(t, err, "Should succeed with Rename policy")

		close(progChan)
		for range progChan {
		}
	})

	t.Run("Copy error", func(t *testing.T) {
		fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
			return nil, fmt.Errorf("read error")
		}
		err := CopyMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Overwrite)
		if err == nil {
			t.Error("Expected error from Copy")
		}
	})

	t.Run("ConflictSkip skips the conflicting file", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{NameStr: "a"}, nil
			}
			return &testutil.MockFileInfo{NameStr: fs.Base(path)}, nil
		}
		copyCalled := false
		fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
			if path == "/dest/a" {
				copyCalled = true
			}
			return testutil.NewMockFile(fs.Base(path), nil), nil
		}

		err := CopyMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Skip)
		testutil.AssertNoError(t, err, "Should succeed with skip")
		testutil.AssertEqual(t, false, copyCalled, "Conflicting file should NOT be copied")
	})

	t.Run("Generic resolver error", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return nil, fmt.Errorf("stat error")
		}
		err := CopyMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Overwrite)
		if err == nil || !strings.Contains(err.Error(), "stat error") {
			t.Errorf("Expected stat error, got %v", err)
		}
	})
}

func TestMoveMultiple_Extra(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return false, nil }
	fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{NameStr: fs.Base(path)}, nil
	}
	fs.LstatFunc = fs.StatFunc
	fs.RenameFunc = func(ctx context.Context, old, new string) error { return nil }
	fs.CreateFunc = func(ctx context.Context, path string) (io.WriteCloser, error) {
		return testutil.NewMockFile(fs.Base(path), nil), nil
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile(fs.Base(path), nil), nil
	}
	fs.RemoveAllFunc = func(ctx context.Context, path string) error { return nil }

	sources := []string{"/src/a"}
	destDir := "/dest"

	t.Run("ValidateWritable error", func(t *testing.T) {
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return true, nil }
		err := MoveMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Overwrite)
		if err == nil {
			t.Error("Expected error when destination is read-only")
		}
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) { return false, nil }
	})

	t.Run("Context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := MoveMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Overwrite)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})

	t.Run("Progress and Rename", func(t *testing.T) {
		progChan := make(chan core.Progress, 10)
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/dest/a" {
				return &testutil.MockFileInfo{NameStr: "a"}, nil
			}
			return nil, os.ErrNotExist
		}
		err := MoveMultiple(ctx, fs, fs, sources, destDir, progChan, conflict.Rename)
		testutil.AssertNoError(t, err, "Should succeed with Rename policy")
	})

	t.Run("Move error", func(t *testing.T) {
		fs.RenameFunc = func(ctx context.Context, old, new string) error {
			return fmt.Errorf("move error")
		}
		err := MoveMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Overwrite)
		if err == nil {
			t.Error("Expected error from Move")
		}
	})

	t.Run("Generic resolver error", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return nil, fmt.Errorf("stat error")
		}
		err := MoveMultiple(ctx, fs, fs, sources, destDir, nil, conflict.Overwrite)
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
