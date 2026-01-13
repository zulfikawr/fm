package ops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files/local"
	"fm/internal/testutil"
)

func TestDeleteMultiple(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	fs := &local.LocalFS{}
	f1 := testutil.CreateTestFile(t, tmpDir, "f1.txt", "c1")
	f2 := testutil.CreateTestFile(t, tmpDir, "f2.txt", "c2")

	err := DeleteMultiple(context.Background(), fs, []string{f1, f2}, false, nil)
	if err != nil {
		t.Fatalf("DeleteMultiple failed: %v", err)
	}

	if _, err := os.Stat(f1); err == nil {
		t.Error("f1 still exists")
	}
	if _, err := os.Stat(f2); err == nil {
		t.Error("f2 still exists")
	}
}

func TestPaste(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	fs := &local.LocalFS{}
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(dstDir, 0755)

	f1 := testutil.CreateTestFile(t, srcDir, "f1.txt", "c1")

	err := Paste(context.Background(), fs, []string{f1}, dstDir, nil)
	if err != nil {
		t.Fatalf("Paste failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dstDir, "f1.txt")); err != nil {
		t.Error("f1.txt not found in destination")
	}

	// Test conflict
	err = Paste(context.Background(), fs, []string{f1}, dstDir, nil)
	if err == nil {
		t.Error("Expected conflict error, got nil")
	}
	if _, ok := err.(*ConflictError); !ok {
		t.Errorf("Expected ConflictError, got %T", err)
	}
}

func TestMoveMultiple(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	fs := &local.LocalFS{}
	srcDir := filepath.Join(tmpDir, "src_move")
	dstDir := filepath.Join(tmpDir, "dst_move")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(dstDir, 0755)

	f1 := testutil.CreateTestFile(t, srcDir, "f1.txt", "c1")

	err := MoveMultiple(context.Background(), fs, []string{f1}, dstDir, nil)
	if err != nil {
		t.Fatalf("MoveMultiple failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dstDir, "f1.txt")); err != nil {
		t.Error("f1.txt not found in destination")
	}
	if _, err := os.Stat(f1); err == nil {
		t.Error("f1.txt still exists in source")
	}
}

func TestCheckAndMarkProcessing(t *testing.T) {
	processing := make(map[string]bool)
	paths := []string{"/p1", "/p2"}

	if !CheckAndMarkProcessing(processing, paths) {
		t.Error("Expected true for new paths")
	}
	if !processing["/p1"] || !processing["/p2"] {
		t.Error("Expected paths to be marked as processing")
	}

	if CheckAndMarkProcessing(processing, []string{"/p1"}) {
		t.Error("Expected false for already processing path")
	}
}
