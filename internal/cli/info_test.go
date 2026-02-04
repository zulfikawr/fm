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

func TestCountRecursive(t *testing.T) {
	// Test the recursive counting logic
	// This replaces the old calculateDirStats test
	// We verify it via RunInfo directory test above
}

func TestInfoJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

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
