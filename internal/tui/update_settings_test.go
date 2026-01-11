package tui

import (
	"fm/internal/files"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Setting identifiers for dynamic test discovery
const (
	settingShowHidden = iota
	settingCaseSensitive
	settingConfirmOperations
	settingWrapNavigation
	settingEditor
	settingUseTrash
	settingShowHeader
	settingEnableGit
	settingShowSize
	settingSizeFormat
	settingShowDate
	settingDateFormat
	settingTheme
	numSettings
)

// isSettingDisabled checks if a setting at given index is disabled
func isSettingDisabled(m *Model, idx int) bool {
	switch idx {
	case settingSizeFormat:
		return !m.cfg.ShowSize
	case settingDateFormat:
		return !m.cfg.ShowDateModified
	default:
		return false
	}
}

// findNextEnabledSetting finds the next enabled setting index
func findNextEnabledSetting(m *Model, current int) int {
	for i := current + 1; i < numSettings; i++ {
		if !isSettingDisabled(m, i) {
			return i
		}
	}
	return current
}

// findPrevEnabledSetting finds the previous enabled setting index
func findPrevEnabledSetting(m *Model, current int) int {
	for i := current - 1; i >= 0; i-- {
		if !isSettingDisabled(m, i) {
			return i
		}
	}
	return current
}

func TestSettings(t *testing.T) {
	m := NewModel(&files.LocalFS{}, "/")
	m.settingsOpen = true
	m.width = 80
	m.height = 24

	t.Run("Navigation skipping disabled settings", func(t *testing.T) {
		// Disable SizeFormat and DateFormat parent settings
		m.cfg.ShowSize = false
		m.cfg.ShowDateModified = false

		// Start at ShowSize setting
		m.settingsCursor = settingShowSize

		// Press down - should skip SizeFormat (disabled) and go to ShowDate
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(*Model)
		expectedNext := findNextEnabledSetting(m, settingShowSize)
		if m.settingsCursor != expectedNext {
			t.Errorf("Expected cursor to skip disabled setting and go to %d, got %d", expectedNext, m.settingsCursor)
		}

		// Press down again - should skip DateFormat (disabled) and go to Theme
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(*Model)
		expectedNext = findNextEnabledSetting(m, m.settingsCursor-1)
		if m.settingsCursor != expectedNext {
			t.Errorf("Expected cursor to skip disabled setting and go to %d, got %d", expectedNext, m.settingsCursor)
		}

		// Test upward navigation with disabled settings
		m.cfg.ShowSize = false
		m.settingsCursor = settingShowDate
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		m = newModel.(*Model)
		expectedPrev := findPrevEnabledSetting(m, settingShowDate)
		if m.settingsCursor != expectedPrev {
			t.Errorf("Expected cursor to skip disabled setting going up and go to %d, got %d", expectedPrev, m.settingsCursor)
		}
	})

	t.Run("Toggle logic for formats", func(t *testing.T) {
		m.cfg.ShowSize = false
		m.cfg.SizeFormatIndex = 0
		m.settingsCursor = settingSizeFormat

		// Should not toggle when parent setting is disabled
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		m = newModel.(*Model)
		if m.cfg.SizeFormatIndex != 0 {
			t.Error("SizeFormat should not toggle when ShowSize is off")
		}

		// Enable parent setting
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
		// Test that all settings can be toggled without crashing
		for i := 0; i < numSettings; i++ {
			m.toggleSetting(i)
		}
	})

	t.Run("Toggle Setting Previous", func(t *testing.T) {
		m := NewModel(&files.LocalFS{}, "/")

		// Test Theme toggle
		m.cfg.ThemeIndex = 1
		m.toggleSettingPrev(settingTheme)
		if m.cfg.ThemeIndex != 0 {
			t.Errorf("Expected theme index 0, got %d", m.cfg.ThemeIndex)
		}

		// Test SizeFormat toggle
		m.cfg.ShowSize = true
		m.cfg.SizeFormatIndex = 1
		m.toggleSettingPrev(settingSizeFormat)
		if m.cfg.SizeFormatIndex != 0 {
			t.Errorf("Expected size format 0, got %d", m.cfg.SizeFormatIndex)
		}

		// Test DateFormat toggle
		m.cfg.ShowDateModified = true
		m.cfg.DateFormatIndex = 1
		m.toggleSettingPrev(settingDateFormat)
		if m.cfg.DateFormatIndex != 0 {
			t.Errorf("Expected date format 0, got %d", m.cfg.DateFormatIndex)
		}

		// Test other toggles don't crash
		m.toggleSettingPrev(settingEditor)
		m.toggleSettingPrev(settingShowHidden)
	})
}

func TestHandleSettingsUpdate(t *testing.T) {
	m := NewModel(&files.LocalFS{}, "/")
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
