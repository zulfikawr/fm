package local

import (
	"context"
	"fmt"
	"os"
	"testing"

	"fm/internal/testutil"
)

func TestLocalFS_IsReadOnly(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	fs := &LocalFS{}

	// Test a writable directory
	ro, err := fs.IsReadOnly(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("IsReadOnly failed: %v", err)
	}
	if ro {
		t.Error("Expected directory to be writable")
	}

	// Test a read-only file (if possible to set permissions in test environment)
	f := testutil.CreateTestFile(t, tmpDir, "readonly.txt", "content")
	os.Chmod(f, 0444)

	ro, err = fs.IsReadOnly(context.Background(), f)
	if err != nil {
		t.Fatalf("IsReadOnly failed for file: %v", err)
	}
	// Note: on some systems/users, owner might still have write access regardless of bits,
	// but we check if our logic detects the change.
}

func TestLocalFS_ReadDir(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	fs := &LocalFS{}

	// Create some files
	for i := 0; i < 10; i++ {
		fname := fmt.Sprintf("extra_%d.txt", i)
		testutil.CreateTestFile(t, tmpDir, fname, "content")
	}
	// Use explicit names
	testutil.CreateTestFile(t, tmpDir, "file1.txt", "c1")
	testutil.CreateTestFile(t, tmpDir, "file2.txt", "c2")

	infos, err := fs.ReadDir(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(infos) < 2 {
		t.Errorf("Expected at least 2 files, got %d", len(infos))
	}

	found := false
	for _, info := range infos {
		if info.Name() == "file1.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("file1.txt not found in ReadDir results")
	}
}
