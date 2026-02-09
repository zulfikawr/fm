package file_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/file"
	"github.com/zulfikawr/fm/internal/tui/messages"
)

func TestFileOps_Basic(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	// Test StartCreate
	file.StartCreate(m)
	if m.Inputs.Mode != tuictx.InputCreate {
		t.Error("expected InputCreate mode")
	}

	// Test PerformCreate
	file.PerformCreate(m, "newfile.txt")
}

func TestFileOps_Rename(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Navigation.FilteredItems = []core.Item{
		{Name: "old.txt", Path: "/test/old.txt"},
	}

	file.StartRename(m)
	if m.Inputs.Mode != tuictx.InputRename {
		t.Error("expected InputRename mode")
	}

	file.PerformRename(m, "new.txt")
}

func TestFileOps_Conflict(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	msg := messages.ConflictMsg{
		Src:    "/src/a",
		Dst:    "/dest/a",
		OpType: "copy",
	}

	file.HandleConflict(m, msg)
	if !m.UI.Confirming {
		t.Error("expected Confirming state")
	}
}

func TestHandleFileKeys(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	item := core.Item{Name: "file1.zip", Path: "/test/file1.zip"}
	m.Navigation.Items = []core.Item{item}
	m.Navigation.FilteredItems = []core.Item{item}
	m.Navigation.Cursor = 0

	tests := []struct {
		key  string
		mode tuictx.InputMode
	}{
		{"y", tuictx.InputNone},
		{"c", tuictx.InputNone},
		{"x", tuictx.InputNone},
		{"r", tuictx.InputRename},
		{"z", tuictx.InputZip},
		{"u", tuictx.InputUnzip},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			m := tuictx.NewModel(fs, "/test")
			item := core.Item{Name: "file1.zip", Path: "/test/file1.zip"}
			m.Navigation.Items = []core.Item{item}
			m.Navigation.FilteredItems = []core.Item{item}
			m.Navigation.Cursor = 0

			file.HandleFileOps(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if tt.mode != tuictx.InputNone && m.Inputs.Mode != tt.mode {
				t.Errorf("key %s: expected mode %v, got %v", tt.key, tt.mode, m.Inputs.Mode)
			}
		})
	}
}

func TestHandleConfirmKeys(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")

	t.Run("Confirm Delete", func(t *testing.T) {
		m.UI.StartConfirming()
		m.Operations.ActionType = "delete"
		file.HandleFileOps(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
		if m.UI.Confirming {
			t.Error("expected Confirming to be false after 'y'")
		}
	})

	t.Run("Cancel Delete", func(t *testing.T) {
		m.UI.StartConfirming()
		m.Operations.ActionType = "delete"
		file.HandleFileOps(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
		if m.UI.Confirming {
			t.Error("expected Confirming to be false after 'n'")
		}
	})
}

func TestPerformRename(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	item := core.Item{Name: "old.txt", Path: "/test/old.txt"}
	m.Navigation.Items = []core.Item{item}
	m.Navigation.FilteredItems = []core.Item{item}
	m.Navigation.Cursor = 0

	cmd := file.PerformRename(m, "new.txt")
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
}

func TestPerformDelete(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	item := core.Item{Name: "file1.txt", Path: "/test/file1.txt"}
	m.Navigation.Items = []core.Item{item}
	m.Navigation.FilteredItems = []core.Item{item}
	m.Navigation.Cursor = 0

	// Mocking GetTargets via item selection
	m.Navigation.Select(item.Path)

	t.Run("Immediate Delete", func(t *testing.T) {
		m.Config.Ops.ConfirmOperations = false
		cmd := file.PerformDelete(m)
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}
	})

	t.Run("Confirm Delete", func(t *testing.T) {
		m.Config.Ops.ConfirmOperations = true
		file.PerformDelete(m)
		if !m.UI.Confirming {
			t.Error("expected Confirming to be true")
		}
	})
}

func TestPerformPaste(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Operations.Clipboard.SetCopy(fs, []string{"/src/file1.txt"})

	t.Run("Immediate Paste", func(t *testing.T) {
		m.Config.Ops.ConfirmOperations = false
		cmd := file.PerformPaste(m)
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}
	})
}

func TestPerformZipUnzip(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	item := core.Item{Name: "file1.txt", Path: "/test/file1.txt"}
	m.Navigation.Items = []core.Item{item}
	m.Navigation.FilteredItems = []core.Item{item}
	m.Navigation.Cursor = 0

	t.Run("PerformZip", func(t *testing.T) {
		cmd := file.PerformZip(m, "archive.zip")
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}
	})

	t.Run("PerformUnzip", func(t *testing.T) {
		zipItem := core.Item{Name: "archive.zip", Path: "/test/archive.zip"}
		m.Navigation.FilteredItems[0] = zipItem
		cmd := file.PerformUnzip(m, "extracted")
		if cmd == nil {
			t.Fatal("expected non-nil command")
		}
	})
}

func TestResolveConflict(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := tuictx.NewModel(fs, "/test")
	m.Operations.Conflict.Set(tuictx.ConflictParams{
		Source:       "/src/a",
		Destination:  "/dest/a",
		PendingItems: []string{"/src/a"},
		IsMove:       false,
		OpType:       "copy",
		LogID:        "log1",
	})

	t.Run("Overwrite", func(t *testing.T) {
		cmd := file.ResolveConflict(m, "overwrite", false)
		if cmd == nil {
			t.Error("expected non-nil command")
		}
	})

	t.Run("Skip", func(t *testing.T) {
		m.Operations.Conflict.Set(tuictx.ConflictParams{
			Source:       "/src/a",
			Destination:  "/dest/a",
			PendingItems: []string{"/src/a"},
			IsMove:       false,
			OpType:       "copy",
			LogID:        "log1",
		})
		cmd := file.ResolveConflict(m, "skip", false)
		if cmd == nil {
			t.Error("expected non-nil command")
		}
	})
}
