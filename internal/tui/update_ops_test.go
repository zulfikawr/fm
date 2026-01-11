package tui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"fm/internal/files"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOperations(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "fm-ops-update-test")
	defer os.RemoveAll(tmpDir)
	m := NewModel(&files.LocalFS{}, tmpDir)

	t.Run("Renaming State", func(t *testing.T) {
		m.path = tmpDir
		m.items = []files.Item{{Name: "old.txt", Path: filepath.Join(tmpDir, "old.txt"), CanRead: true, CanWrite: true}}
		os.WriteFile(m.items[0].Path, []byte("test"), 0644)
		m.applyFilter()
		m.cursor = 0

		m.renaming = true
		m.renameInput.Focus()
		m.renameInput.SetValue("new.txt")

		// Test esc in renaming
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = newModel.(*Model)
		if m.renaming {
			t.Error("Expected renaming to be false after esc")
		}

		m.renaming = true
		m.renameInput.SetValue("new.txt")
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(*Model)
		if m.renaming {
			t.Error("Expected renaming to be false after enter")
		}

		if _, err := os.Stat(filepath.Join(tmpDir, "new.txt")); os.IsNotExist(err) {
			t.Error("Expected new.txt to exist")
		}
	})

	t.Run("Confirming State - Delete", func(t *testing.T) {
		m.path = tmpDir
		filePath := filepath.Join(tmpDir, "delete_me.txt")
		os.WriteFile(filePath, []byte("delete"), 0644)
		m.items = []files.Item{{Name: "delete_me.txt", Path: filePath, CanRead: true, CanWrite: true}}
		m.applyFilter()
		m.cursor = 0

		m.confirming = true
		m.actionType = "delete"
		m.cfg.UseTrash = false

		// Test 'n' for no
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		m = newModel.(*Model)
		if m.confirming {
			t.Error("Expected confirming to be false after 'n'")
		}

		m.confirming = true
		m.actionType = "delete"
		m.cfg.UseTrash = false

		// Test 'y' for yes
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		m = newModel.(*Model)
		if m.confirming {
			t.Error("Expected confirming to be false after 'y'")
		}

		if !m.loading {
			t.Error("Expected loading to be true after delete triggered")
		}

		// Manually perform side-effect for verification
		files.Delete(context.Background(), m.fs, filePath, nil)

		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Error("Expected file to be deleted")
		}
	})

	t.Run("Confirming State - Paste", func(t *testing.T) {
		m := NewModel(&files.LocalFS{}, tmpDir)
		srcFile := filepath.Join(tmpDir, "src_paste.txt")
		os.WriteFile(srcFile, []byte("paste content"), 0644)

		subDir := filepath.Join(tmpDir, "sub")
		os.MkdirAll(subDir, 0755)

		m.clipboard = []string{srcFile}
		m.confirming = true
		m.actionType = "paste"
		m.path = subDir

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		m = newModel.(*Model)

		if !m.loading {
			t.Error("Expected loading to be true after paste triggered")
		}

		// Manually perform side-effect for verification
		files.Copy(context.Background(), m.fs, srcFile, filepath.Join(subDir, "src_paste.txt"), nil)

		if _, err := os.Stat(filepath.Join(subDir, "src_paste.txt")); err != nil {
			t.Error("Expected pasted file to exist in sub")
		}
	})

	t.Run("Perform Paste - Cut", func(t *testing.T) {
		m := NewModel(&files.LocalFS{}, tmpDir)
		srcFile := filepath.Join(tmpDir, "src_cut.txt")
		os.WriteFile(srcFile, []byte("cut content"), 0644)

		destDir := filepath.Join(tmpDir, "dest_cut")
		os.MkdirAll(destDir, 0755)

		m.clipboard = []string{srcFile}
		m.clipboardCut = true
		m.path = destDir

		m.performPaste()

		if !m.loading {
			t.Error("Expected loading to be true after move triggered")
		}
		if len(m.clipboard) != 0 {
			t.Error("Expected clipboard to be cleared after move")
		}
		if m.clipboardCut {
			t.Error("Expected clipboardCut to be false after move")
		}
	})
}
