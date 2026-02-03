package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/files/local"
)

func TestRunInfo(t *testing.T) {
	// Create a temporary test directory
	tmpDir := t.TempDir()

	// Create test files
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	testSubDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(testSubDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("File info", func(t *testing.T) {
		opts := InfoOptions{
			Path: testFile,
			JSON: false,
		}

		err := RunInfo(opts)
		if err != nil {
			t.Errorf("RunInfo failed: %v", err)
		}
	})

	t.Run("Directory info", func(t *testing.T) {
		opts := InfoOptions{
			Path: tmpDir,
			JSON: false,
		}

		err := RunInfo(opts)
		if err != nil {
			t.Errorf("RunInfo failed: %v", err)
		}
	})

	t.Run("JSON output", func(t *testing.T) {
		opts := InfoOptions{
			Path: testFile,
			JSON: true,
		}

		// Capture stdout would require more complex setup
		// For now, just verify it doesn't error
		err := RunInfo(opts)
		if err != nil {
			t.Errorf("RunInfo with JSON failed: %v", err)
		}
	})

	t.Run("Tree view", func(t *testing.T) {
		opts := InfoOptions{
			Path:      tmpDir,
			Tree:      true,
			TreeDepth: 2,
		}

		err := RunInfo(opts)
		if err != nil {
			t.Errorf("RunInfo with tree failed: %v", err)
		}
	})

	t.Run("Nonexistent path", func(t *testing.T) {
		opts := InfoOptions{
			Path: filepath.Join(tmpDir, "nonexistent"),
		}

		err := RunInfo(opts)
		if err == nil {
			t.Error("Expected error for nonexistent path")
		}
	})
}

func TestCalculateDirStats(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	subdir := filepath.Join(tmpDir, "subdir")
	file3 := filepath.Join(subdir, "file3.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file3, []byte("content3"), 0644); err != nil {
		t.Fatal(err)
	}

	fs := local.NewLocalFS()
	ctx := context.Background()

	stats, err := calculateDirStats(ctx, fs, tmpDir)
	if err != nil {
		t.Fatalf("calculateDirStats failed: %v", err)
	}

	if stats.FileCount != 3 {
		t.Errorf("Expected 3 files, got %d", stats.FileCount)
	}

	if stats.DirectoryCount != 1 {
		t.Errorf("Expected 1 directory, got %d", stats.DirectoryCount)
	}

	if stats.TotalSize == 0 {
		t.Error("Expected non-zero total size")
	}
}

func TestBuildTree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "file2.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	fs := local.NewLocalFS()
	ctx := context.Background()

	t.Run("Depth 0 (unlimited)", func(t *testing.T) {
		tree, err := buildTree(ctx, fs, tmpDir, 0, 0)
		if err != nil {
			t.Fatalf("buildTree failed: %v", err)
		}

		if !tree.IsDir {
			t.Error("Root should be a directory")
		}

		if len(tree.Children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(tree.Children))
		}
	})

	t.Run("Depth 1", func(t *testing.T) {
		tree, err := buildTree(ctx, fs, tmpDir, 0, 1)
		if err != nil {
			t.Fatalf("buildTree failed: %v", err)
		}

		// Should have children but subdirectory children should not be loaded
		hasSubdir := false
		for _, child := range tree.Children {
			if child.IsDir {
				hasSubdir = true
				if len(child.Children) > 0 {
					t.Error("Subdirectory should have no children at depth 1")
				}
			}
		}

		if !hasSubdir {
			t.Error("Expected to find subdirectory")
		}
	})
}

func TestCalculateGitStats(t *testing.T) {
	// Git status format is XY where:
	// X = index status (first character)
	// Y = working tree status (second character)
	statuses := map[string]string{
		"file1.txt": "M ", // M in index (staged), unmodified in working tree
		"file2.txt": "A ", // A in index (staged addition), unmodified in working tree
		"file3.txt": "??", // Untracked (not in index or tree)
		"file4.txt": " D", // Unmodified in index, deleted in working tree
		"file5.txt": "MM", // M in index (staged), M in working tree
	}

	stats := calculateGitStats(statuses)

	// Modified in working tree: file5 (MM) only
	// file4 ( D) is deleted, not modified
	if stats.Modified != 1 {
		t.Errorf("Expected 1 modified in working tree, got %d", stats.Modified)
	}

	// Added: file2 (A ) - added to index
	if stats.Added != 1 {
		t.Errorf("Expected 1 added, got %d", stats.Added)
	}

	// Untracked: file3 (??) - not in git
	if stats.Untracked != 1 {
		t.Errorf("Expected 1 untracked, got %d", stats.Untracked)
	}

	// Deleted: file4 ( D) - deleted in working tree
	if stats.Deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", stats.Deleted)
	}

	// Staged: file1 (M ), file2 (A ), file5 (MM)
	// All have changes in the index (first character)
	if stats.Staged != 3 {
		t.Errorf("Expected 3 staged, got %d", stats.Staged)
	}
}

func TestInfoJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// We can't easily capture stdout, but we can test the JSON structure
	// by building the result manually
	fs := local.NewLocalFS()
	ctx := context.Background()

	info, err := fs.Stat(ctx, testFile)
	if err != nil {
		t.Fatal(err)
	}

	result := &InfoResult{
		Path:        testFile,
		Type:        getTypeString(info),
		Size:        info.Size(),
		Permissions: info.Mode().String(),
		CanRead:     canRead(info),
		CanWrite:    canWrite(info),
		InGitRepo:   false,
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal result: %v", err)
	}

	// Verify we can unmarshal it back
	var decoded InfoResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}

	if decoded.Path != testFile {
		t.Errorf("Path mismatch: got %s, want %s", decoded.Path, testFile)
	}

	if decoded.Type != "file" {
		t.Errorf("Type mismatch: got %s, want file", decoded.Type)
	}
}
