package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSettings(t *testing.T) {
	m := NewModel("/")
	m.settingsOpen = true
	m.width = 80
	m.height = 24

	// New indices:
	// 0: ShowHidden, 1: CaseSensitive, 2: Confirmations, 3: Wrap, 4: Editor
	// 5: ShowHeader, 6: EnableGit, 7: ShowSize, 8: SizeFormat, 9: ShowDate, 10: DateFormat, 11: Theme

	t.Run("Navigation skipping disabled settings", func(t *testing.T) {
		m.cfg.ShowSize = false
		m.cfg.ShowDateModified = false
		m.settingsCursor = 7 // On ShowSize (off)
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(*Model)
		// Should skip 8 and go to 9
		if m.settingsCursor != 9 {
			t.Errorf("Expected cursor to skip 8 and go to 9, got %d", m.settingsCursor)
		}

		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(*Model)
		// Should skip 10 and go to 11
		if m.settingsCursor != 11 {
			t.Errorf("Expected cursor to skip 10 and go to 11, got %d", m.settingsCursor)
		}
	})

	t.Run("Toggle logic for formats", func(t *testing.T) {
		m.cfg.ShowSize = false
		m.cfg.SizeFormatIndex = 0
		m.settingsCursor = 8 // On SizeFormat

		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)
		if m.cfg.SizeFormatIndex != 0 {
			t.Error("SizeFormat should not toggle when ShowSize is off")
		}

		m.cfg.ShowSize = true
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)
		if m.cfg.SizeFormatIndex == 0 {
			t.Error("SizeFormat should toggle when ShowSize is true")
		}
	})

	t.Run("Toggle Setting All", func(t *testing.T) {
		m.cfg.ShowSize = true
		m.cfg.ShowDateModified = true
		for i := 0; i <= 11; i++ {
			m.toggleSetting(i)
		}
	})

	t.Run("Toggle Setting Previous", func(t *testing.T) {
		m := NewModel("/")
		m.cfg.ThemeIndex = 1
		m.toggleSettingPrev(11) // Theme
		if m.cfg.ThemeIndex != 0 {
			t.Errorf("Expected theme index 0, got %d", m.cfg.ThemeIndex)
		}

		m.cfg.ShowSize = true
		m.cfg.SizeFormatIndex = 1
		m.toggleSettingPrev(8) // Size format
		if m.cfg.SizeFormatIndex != 0 {
			t.Errorf("Expected size format 0, got %d", m.cfg.SizeFormatIndex)
		}

		m.cfg.ShowDateModified = true
		m.cfg.DateFormatIndex = 1
		m.toggleSettingPrev(10) // Date format
		if m.cfg.DateFormatIndex != 0 {
			t.Errorf("Expected date format 0, got %d", m.cfg.DateFormatIndex)
		}

		m.toggleSettingPrev(4) // Editor prev
		m.toggleSettingPrev(0)
	})
}

func TestHandleSettingsUpdate(t *testing.T) {
	m := NewModel("/")
	m.settingsOpen = true

	t.Run("Esc to close", func(t *testing.T) {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if newModel.(*Model).settingsOpen {
			t.Error("Expected settings to be closed")
		}
	})

	m.settingsOpen = true
	t.Run("Navigation", func(t *testing.T) {
		m.settingsCursor = 0
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if m.settingsCursor != 1 {
			t.Errorf("Expected cursor 1, got %d", m.settingsCursor)
		}
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if m.settingsCursor != 0 {
			t.Errorf("Expected cursor 0, got %d", m.settingsCursor)
		}
	})

	t.Run("Toggle prev", func(t *testing.T) {
		m.settingsCursor = 11 // Theme
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	})
}
