package tui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

type MockFS struct {
	files.FileSystem
	RemoveAllFunc func(path string) error
}

func (m *MockFS) RemoveAll(ctx context.Context, path string) error {
	if m.RemoveAllFunc != nil {
		return m.RemoveAllFunc(path)
	}
	return m.FileSystem.RemoveAll(ctx, path)
}

func TestCommands(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-commands-test")
	defer os.RemoveAll(tmpDir)
	m := NewModel(&files.LocalFS{}, tmpDir)

	t.Run("LoadedItemsMsg", func(t *testing.T) {
		items := []files.Item{{Name: "test.txt", IsDir: false}}
		msg := LoadedItemsMsg{
			Path:  m.path,
			Items: items,
		}
		newModel, _ := m.Update(msg)
		m = newModel.(*Model)
		if len(m.items) != 1 || m.items[0].Name != "test.txt" {
			t.Errorf("Expected 1 item 'test.txt', got %d items", len(m.items))
		}
		if m.loading {
			t.Error("Expected loading to be false after LoadedItemsMsg")
		}
	})

	t.Run("Init", func(t *testing.T) {
		cmd := m.Init()
		if cmd == nil {
			t.Error("Expected non-nil Init command")
		}
	})

	t.Run("Watch Event Msg", func(t *testing.T) {
		msg := WatchEventMsg{}
		m.Update(msg)
	})

	t.Run("Paste and Move Commands", func(t *testing.T) {
		src := filepath.Join(tmpDir, "cmd_src.txt")
		os.WriteFile(src, []byte("content"), 0644)
		dest := filepath.Join(tmpDir, "cmd_dest")
		os.MkdirAll(dest, 0755)

		// Test pasteItems command execution
		cmd := pasteItems(m.fs, []string{src}, dest)
		batchMsg := cmd().(tea.BatchMsg)
		// The second command in batch is the actual operation
		msg := batchMsg[1]()
		if _, ok := msg.(OperationFinishedMsg); !ok {
			t.Errorf("Expected OperationFinishedMsg on success, got %T (%v)", msg, msg)
		}

		// Delete original src before move to avoid conflict
		os.Remove(src)

		// Test moveItems command execution
		cmd = moveItems(m.fs, []string{filepath.Join(dest, "cmd_src.txt")}, tmpDir)
		batchMsg = cmd().(tea.BatchMsg)
		msg = batchMsg[1]()
		if _, ok := msg.(OperationFinishedMsg); !ok {
			t.Errorf("Expected OperationFinishedMsg on success, got %T (%v)", msg, msg)
		}
	})

	t.Run("Delete Items Command Detailed", func(t *testing.T) {
		f1 := filepath.Join(tmpDir, "to_del1.txt")
		os.WriteFile(f1, []byte("1"), 0644)

		// Success case
		cmd := deleteItems(m.fs, []string{f1}, false)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		if _, ok := msg.(OperationFinishedMsg); !ok {
			t.Errorf("Expected OperationFinishedMsg on success, got %T (%v)", msg, msg)
		}
		if _, err := os.Stat(f1); !os.IsNotExist(err) {
			t.Error("File should have been deleted")
		}

		// Error case
		mock := &MockFS{
			FileSystem: m.fs,
			RemoveAllFunc: func(path string) error {
				return os.ErrPermission
			},
		}
		cmd = deleteItems(mock, []string{"any-path"}, false)
		batchMsg = cmd().(tea.BatchMsg)
		msg = batchMsg[1]()
		if _, ok := msg.(errMsg); !ok {
			t.Errorf("Expected errMsg on failure, got %T", msg)
		}
	})

	t.Run("Reload Command", func(t *testing.T) {
		cmd := m.reload()
		msg := cmd()
		if _, ok := msg.(LoadedItemsMsg); !ok {
			t.Errorf("Expected LoadedItemsMsg, got %T", msg)
		}
	})

	t.Run("CalculateDirSize Command", func(t *testing.T) {
		cmd := calculateDirSize(m.fs, tmpDir)
		msg := cmd()
		if _, ok := msg.(DirSizeMsg); !ok {
			t.Errorf("Expected DirSizeMsg, got %T", msg)
		}
	})

	t.Run("FetchGitStatus Command", func(t *testing.T) {
		cmd := fetchGitStatus(m.fs, tmpDir)
		msg := cmd()
		if _, ok := msg.(GitStatusMsg); !ok {
			t.Errorf("Expected GitStatusMsg, got %T", msg)
		}
	})

	t.Run("Watcher Commands", func(t *testing.T) {
		// Initialize watcher
		m.restartWatcher()
		if m.watcher != nil {
			m.watcher.Close()
		}

		// Test watchDir - should return WatcherClosedMsg since we closed it
		cmd := m.watchDir()
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(WatcherClosedMsg); !ok {
				t.Errorf("Expected WatcherClosedMsg, got %T", msg)
			}
		}

		// Test restartWatcher
		cmd = m.restartWatcher()
		if cmd == nil {
			t.Error("Expected non-nil restartWatcher command")
		}

		// Test WatcherErrorMsg manually
		msg := WatcherErrorMsg{Err: os.ErrPermission}
		m.Update(msg)

		// Test WatcherClosedMsg manually
		msg2 := WatcherClosedMsg{}
		m.Update(msg2)
	})

	t.Run("Delete Items with Trash", func(t *testing.T) {
		f1 := filepath.Join(tmpDir, "to_trash.txt")
		os.WriteFile(f1, []byte("trash me"), 0644)

		// Trash might fail on some systems/environments, but we can at least try to call it
		cmd := deleteItems(m.fs, []string{f1}, true)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		// It might be errMsg if trash is not supported, or OperationFinishedMsg
		_ = msg
	})

	t.Run("Paste and Move Error Cases", func(t *testing.T) {
		// Since files.Copy is a package-level function, we can't easily mock it
		// without changing how it's called.
		// However, we can use a non-existent source to trigger error.

		cmd := pasteItems(m.fs, []string{"non-existent"}, tmpDir)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		if _, ok := msg.(errMsg); !ok {
			t.Errorf("Expected errMsg for non-existent source, got %T", msg)
		}

		cmd = moveItems(m.fs, []string{"non-existent"}, tmpDir)
		batchMsg = cmd().(tea.BatchMsg)
		msg = batchMsg[1]()
		if _, ok := msg.(errMsg); !ok {
			t.Errorf("Expected errMsg for non-existent source, got %T", msg)
		}
	})

	t.Run("ClearMessage Command", func(t *testing.T) {
		cmd := clearMessage()
		if cmd == nil {
			t.Fatal("Expected non-nil command")
		}
		// We can't easily test the Tick result without waiting, but we can verify it's a Tick.
	})

	t.Run("ListenToProgress Command", func(t *testing.T) {
		progChan := make(chan files.Progress, 1)
		progChan <- files.Progress{Percent: 0.5, Label: "Testing"}
		cmd := listenToProgress(progChan)
		msg := cmd()
		pMsg, ok := msg.(ProgressMsg)
		if !ok {
			t.Errorf("Expected ProgressMsg, got %T", msg)
		}
		if pMsg.Percent != 0.5 {
			t.Errorf("Expected 0.5, got %f", pMsg.Percent)
		}

		close(progChan)
		msg = cmd()
		if msg != nil {
			t.Error("Expected nil msg when channel closed")
		}
	})

	t.Run("OverwriteItem Command", func(t *testing.T) {
		src := filepath.Join(tmpDir, "ov_src.txt")
		dst := filepath.Join(tmpDir, "ov_dst.txt")
		os.WriteFile(src, []byte("src"), 0644)

		cmd := overwriteItem(m.fs, src, dst, false)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		if _, ok := msg.(OperationFinishedMsg); !ok {
			t.Errorf("Expected OperationFinishedMsg, got %T", msg)
		}
	})

	t.Run("Paste and Move Conflict", func(t *testing.T) {
		src := filepath.Join(tmpDir, "conf_src.txt")
		dst := filepath.Join(tmpDir, "conf_dst.txt")
		os.WriteFile(src, []byte("src"), 0644)
		os.WriteFile(dst, []byte("dst"), 0644)

		// Test paste conflict
		cmd := pasteItems(m.fs, []string{src}, tmpDir)
		batchMsg := cmd().(tea.BatchMsg)
		msg := batchMsg[1]()
		if cMsg, ok := msg.(conflictMsg); !ok {
			t.Errorf("Expected conflictMsg, got %T", msg)
		} else if cMsg.IsMove {
			t.Error("Expected IsMove to be false")
		}

		// Test move conflict
		cmd = moveItems(m.fs, []string{src}, tmpDir)
		batchMsg = cmd().(tea.BatchMsg)
		msg = batchMsg[1]()
		if cMsg, ok := msg.(conflictMsg); !ok {
			t.Errorf("Expected conflictMsg, got %T", msg)
		} else if !cMsg.IsMove {
			t.Error("Expected IsMove to be true")
		}
	})
}
