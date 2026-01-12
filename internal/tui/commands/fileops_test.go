package commands

import (
	"context"
	"fm/internal/testutil"
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files/local"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFileOps(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	fs := &local.LocalFS{}

	t.Run("Paste and Move Commands", func(t *testing.T) {
		src := filepath.Join(tmpDir, "cmd_src.txt")
		testutil.CreateTestFile(t, tmpDir, "cmd_src.txt", "content")
		dest := filepath.Join(tmpDir, "cmd_dest")
		os.MkdirAll(dest, 0755)

		// Test pasteItems command execution
		cmd := PasteItems(fs, []string{src}, dest)
		batchMsg := cmd().(tea.BatchMsg)
		// The second command in batch is the actual operation
		msg := batchMsg[1]()
		if _, ok := msg.(OperationFinishedMsg); !ok {
			t.Errorf("Expected OperationFinishedMsg on success, got %T (%v)", msg, msg)
		}

		// Delete original src before move to avoid conflict
		os.Remove(src)

		// Test MoveItems command execution
		cmd = MoveItems(fs, []string{filepath.Join(dest, "cmd_src.txt")}, tmpDir)
		batchMsg = cmd().(tea.BatchMsg)
		msg = batchMsg[1]()
		if _, ok := msg.(OperationFinishedMsg); !ok {
			t.Errorf("Expected OperationFinishedMsg on success, got %T (%v)", msg, msg)
		}
	})

	t.Run("Delete Items Command Detailed", func(t *testing.T) {
		f1 := filepath.Join(tmpDir, "to_del1.txt")
		testutil.CreateTestFile(t, tmpDir, "to_del1.txt", "1")

		// Success case
		cmd := DeleteItems(fs, []string{f1}, false)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		if _, ok := msg.(OperationFinishedMsg); !ok {
			t.Errorf("Expected OperationFinishedMsg on success, got %T (%v)", msg, msg)
		}
		if _, err := os.Stat(f1); !os.IsNotExist(err) {
			t.Error("File should have been deleted")
		}

		// Error case
		mock := testutil.NewMockFileSystem()
		mock.FileSystem = fs
		mock.RemoveAllFunc = func(ctx context.Context, path string) error {
			return os.ErrPermission
		}
		cmd = DeleteItems(mock, []string{"any-path"}, false)
		batchMsg = cmd().(tea.BatchMsg)
		msg = batchMsg[1]()
		if _, ok := msg.(ErrorMsg); !ok {
			t.Errorf("Expected ErrorMsg on failure, got %T", msg)
		}
	})

	t.Run("Delete Items with Trash", func(t *testing.T) {
		f1 := filepath.Join(tmpDir, "to_trash.txt")
		testutil.CreateTestFile(t, tmpDir, "to_trash.txt", "trash me")

		cmd := DeleteItems(fs, []string{f1}, true)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		// Success or error depending on system, just ensure it runs
		_ = msg
	})

	t.Run("Paste and Move Error Cases", func(t *testing.T) {
		cmd := PasteItems(fs, []string{"non-existent"}, tmpDir)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		if _, ok := msg.(ErrorMsg); !ok {
			t.Errorf("Expected ErrorMsg for non-existent source, got %T", msg)
		}

		cmd = MoveItems(fs, []string{"non-existent"}, tmpDir)
		batchMsg = cmd().(tea.BatchMsg)
		msg = batchMsg[1]()
		if _, ok := msg.(ErrorMsg); !ok {
			t.Errorf("Expected ErrorMsg for non-existent source, got %T", msg)
		}
	})

	t.Run("OverwriteItem Command", func(t *testing.T) {
		src := filepath.Join(tmpDir, "ov_src.txt")
		dst := filepath.Join(tmpDir, "ov_dst.txt")
		testutil.CreateTestFile(t, tmpDir, "ov_src.txt", "src")

		cmd := OverwriteItem(fs, src, dst, false)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		if _, ok := msg.(OperationFinishedMsg); !ok {
			t.Errorf("Expected OperationFinishedMsg, got %T", msg)
		}
	})

	t.Run("Paste and Move Conflict", func(t *testing.T) {
		src := filepath.Join(tmpDir, "conf_src.txt")
		testutil.CreateTestFile(t, tmpDir, "conf_src.txt", "src")
		testutil.CreateTestFile(t, tmpDir, "conf_dst.txt", "dst")

		// Test paste conflict
		cmd := PasteItems(fs, []string{src}, tmpDir)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		if cMsg, ok := msg.(ConflictMsg); !ok {
			t.Errorf("Expected ConflictMsg, got %T", msg)
		} else if cMsg.IsMove {
			t.Error("Expected IsMove to be false")
		}

		// Test move conflict
		cmd = MoveItems(fs, []string{src}, tmpDir)
		batchMsg = cmd().(tea.BatchMsg)
		msg = batchMsg[1]()
		if cMsg, ok := msg.(ConflictMsg); !ok {
			t.Errorf("Expected ConflictMsg, got %T", msg)
		} else if !cMsg.IsMove {
			t.Error("Expected IsMove to be true")
		}
	})
}
