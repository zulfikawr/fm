package trash

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zulfikawr/fm/internal/files/local"
)

func setupTestTrash(t *testing.T) (*Manager, string, func()) {
	t.Helper()

	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "fm-trash-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	// Override home dir for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	fs := local.NewLocalFS()
	manager, err := NewManager(fs)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("create manager: %v", err)
	}

	cleanup := func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(tmpDir)
	}

	return manager, tmpDir, cleanup
}

func TestNewManager(t *testing.T) {
	manager, tmpDir, cleanup := setupTestTrash(t)
	defer cleanup()

	if tmpDir == "" {
		t.Fatal("Expected non-empty tmpDir")
	}

	// Check directories were created
	if info, err := os.Stat(manager.trashDir); os.IsNotExist(err) {
		t.Errorf("trash dir not created (info: %+v)", info)
	}
	if info, err := os.Stat(manager.filesDir); os.IsNotExist(err) {
		t.Errorf("files dir not created (info: %+v)", info)
	}
	if info, err := os.Stat(manager.infoDir); os.IsNotExist(err) {
		t.Errorf("info dir not created (info: %+v)", info)
	}
}

func TestMoveToTrash(t *testing.T) {
	manager, tmpDir, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	// Move to trash
	ctx := context.Background()
	if err := manager.MoveToTrash(ctx, testFile); err != nil {
		t.Fatalf("move to trash: %v", err)
	}

	// Check original file is gone
	if info, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("original file still exists (info: %+v, error: %v)", info, err)
	}

	// Check file exists in trash
	items, err := manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].OriginalPath != testFile {
		t.Errorf("wrong original path: got %s, want %s", items[0].OriginalPath, testFile)
	}
}

func TestMoveToTrashDirectory(t *testing.T) {
	manager, tmpDir, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create test directory with files
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("create test dir: %v", err)
	}
	testFile := filepath.Join(testDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	// Move to trash
	ctx := context.Background()
	if err := manager.MoveToTrash(ctx, testDir); err != nil {
		t.Fatalf("move to trash: %v", err)
	}

	// Check directory is gone
	if info, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Errorf("original directory still exists (info: %+v, error: %v)", info, err)
	}

	// Check metadata
	items, err := manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if !items[0].IsDirectory {
		t.Error("item should be marked as directory")
	}
}

func TestRestore(t *testing.T) {
	manager, tmpDir, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create and trash a file
	testFile := filepath.Join(tmpDir, "restore-test.txt")
	content := []byte("restore me")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	ctx := context.Background()
	if err := manager.MoveToTrash(ctx, testFile); err != nil {
		t.Fatalf("move to trash: %v", err)
	}

	// Get trashed name
	items, err := manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	// Restore
	if err := manager.Restore(ctx, items[0].TrashedName); err != nil {
		t.Fatalf("restore: %v", err)
	}

	// Check file is back
	restored, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(restored) != string(content) {
		t.Errorf("content mismatch: got %s, want %s", restored, content)
	}

	// Check trash is empty
	items, err = manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("trash should be empty, got %d items", len(items))
	}
}

func TestDelete(t *testing.T) {
	manager, tmpDir, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create and trash a file
	testFile := filepath.Join(tmpDir, "delete-test.txt")
	if err := os.WriteFile(testFile, []byte("delete me"), 0644); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	ctx := context.Background()
	if err := manager.MoveToTrash(ctx, testFile); err != nil {
		t.Fatalf("move to trash: %v", err)
	}

	// Get trashed name
	items, err := manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	trashedName := items[0].TrashedName

	// Delete permanently
	if err := manager.Delete(ctx, trashedName); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Check trash is empty
	items, err = manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("trash should be empty, got %d items", len(items))
	}

	// Check file doesn't exist in trash
	trashedPath := filepath.Join(manager.filesDir, trashedName)
	if info, err := os.Stat(trashedPath); !os.IsNotExist(err) {
		t.Errorf("trashed file still exists (info: %+v, error: %v)", info, err)
	}
}

func TestEmpty(t *testing.T) {
	manager, tmpDir, cleanup := setupTestTrash(t)
	defer cleanup()

	ctx := context.Background()

	// Create and trash multiple files
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("create test file: %v", err)
		}
		if err := manager.MoveToTrash(ctx, testFile); err != nil {
			t.Fatalf("move to trash: %v", err)
		}
	}

	// Check trash has items
	items, err := manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Empty trash
	if err := manager.Empty(ctx); err != nil {
		t.Fatalf("empty trash: %v", err)
	}

	// Check trash is empty
	items, err = manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("trash should be empty, got %d items", len(items))
	}
}

func TestMetadataPreservation(t *testing.T) {
	manager, tmpDir, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create test file with specific permissions
	testFile := filepath.Join(tmpDir, "metadata-test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Fatalf("create test file: %v", err)
	}

	// Move to trash
	ctx := context.Background()
	if err := manager.MoveToTrash(ctx, testFile); err != nil {
		t.Fatalf("move to trash: %v", err)
	}

	// Get metadata
	items, err := manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	info, err := manager.GetInfo(items[0].TrashedName)
	if err != nil {
		t.Fatalf("get info: %v", err)
	}

	// Check metadata
	if info.OriginalPath != testFile {
		t.Errorf("wrong original path: got %s, want %s", info.OriginalPath, testFile)
	}
	if info.Version != MetadataVersion {
		t.Errorf("wrong version: got %d, want %d", info.Version, MetadataVersion)
	}
	if time.Since(info.DeletionTime) > time.Minute {
		t.Error("deletion time too old")
	}
}

func TestRecoverInterruptedDeletions(t *testing.T) {
	manager, tmpDir, cleanup := setupTestTrash(t)
	defer cleanup()

	ctx := context.Background()

	// Create and trash a file
	testFile := filepath.Join(tmpDir, "interrupted-test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("create test file: %v", err)
	}
	if err := manager.MoveToTrash(ctx, testFile); err != nil {
		t.Fatalf("move to trash: %v", err)
	}

	// Get trashed name
	items, err := manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	trashedName := items[0].TrashedName

	// Simulate interrupted deletion by creating marker
	markerPath := filepath.Join(manager.filesDir, trashedName+".deleting")
	if err := os.WriteFile(markerPath, []byte{}, 0644); err != nil {
		t.Fatalf("create marker: %v", err)
	}

	// Recover
	if err := manager.RecoverInterruptedDeletions(ctx); err != nil {
		t.Fatalf("recover: %v", err)
	}

	// Check marker is gone
	if info, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Errorf("marker still exists (info: %+v, error: %v)", info, err)
	}

	// Check trash is empty (item was deleted)
	items, err = manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("trash should be empty, got %d items", len(items))
	}
}

func TestNameCollisionPrevention(t *testing.T) {
	manager, tmpDir, cleanup := setupTestTrash(t)
	defer cleanup()

	ctx := context.Background()

	// Create two files with same name in different locations
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("create dir1: %v", err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatalf("create dir2: %v", err)
	}

	file1 := filepath.Join(dir1, "same.txt")
	file2 := filepath.Join(dir2, "same.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("create file2: %v", err)
	}

	// Trash both
	if err := manager.MoveToTrash(ctx, file1); err != nil {
		t.Fatalf("move file1 to trash: %v", err)
	}
	if err := manager.MoveToTrash(ctx, file2); err != nil {
		t.Fatalf("move file2 to trash: %v", err)
	}

	// Check both are in trash with different names
	items, err := manager.List()
	if err != nil {
		t.Fatalf("list trash: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].TrashedName == items[1].TrashedName {
		t.Error("trashed names should be different")
	}
}
