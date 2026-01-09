package tui

import (
	"os"
	"testing"

	"filemanager/internal/files"
)

func TestCommands(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-commands-test")
	defer os.RemoveAll(tmpDir)
	m := NewModel(tmpDir)

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
}
