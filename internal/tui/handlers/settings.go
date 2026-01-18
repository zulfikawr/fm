package handlers

import (
	"fm/internal/config"
	"fm/internal/constants"
	"fm/internal/files/format"
	tui_context "fm/internal/tui/context"
	"fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleSettings handles settings-related messages
func HandleSettings(m *tui_context.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.UI.SettingsOpen {
			return handleSettingsKeys(m, msg)
		}
	}
	return nil
}

func handleSettingsKeys(m *tui_context.Model, msg tea.KeyMsg) tea.Cmd {
	// Total selectable items across all groups
	// File Ops (6) + Display (6) + Appearance (1) + Keybindings (24) = 37
	totalItems := 37
	var reload bool

	switch msg.String() {
	case "up", "k":
		if m.Settings.Cursor > 0 {
			m.Settings.Cursor--
		}
	case "down", "j":
		if m.Settings.Cursor < totalItems-1 {
			m.Settings.Cursor++
		}
	case "enter", "right", "l", " ":
		reload = toggleSetting(m.Settings.Cursor, m)
	case "left", "h":
		reload = toggleSettingPrev(m.Settings.Cursor, m)
	case "r":
		m.Operations.ActionType = constants.ActionResetSettings
		m.UI.StartConfirming()
	case "esc", "q":
		m.UI.ToggleSettings()
	}

	m.Settings.Offset = scrollSettings(m)

	if reload {
		return Reload(m, false)
	}
	return nil
}

func scrollSettings(m *tui_context.Model) int {
	cursor := m.Settings.Cursor
	offset := m.Settings.Offset
	height := m.Display.ViewportHeight

	// Map selectable item index to actual row index (including headers)
	// Group 1 Header (row 1)
	//   Item 0-5 (rows 2-7)
	// Group 2 Header (row 9)
	//   Item 6-11 (rows 10-15)
	// Group 3 Header (row 17)
	//   Item 12 (row 18)
	// Group 4 Header (row 20)
	//   Item 13-31 (rows 21-39)

	rowIdx := 0
	if cursor <= 5 {
		rowIdx = cursor + 2 // Group 1 header + empty line
	} else if cursor <= 11 {
		rowIdx = cursor + 4 // + previous empty lines and headers
	} else if cursor <= 12 {
		rowIdx = cursor + 6
	} else {
		rowIdx = cursor + 8
	}

	if rowIdx < offset {
		// When scrolling up, we want to see the header of the group if we are at the top item
		newOffset := rowIdx
		if cursor == 0 || cursor == 6 || cursor == 12 || cursor == 13 {
			newOffset -= 2 // Show header and spacing
		}
		if newOffset < 0 {
			newOffset = 0
		}
		return newOffset
	}

	if rowIdx >= offset+height {
		return rowIdx - height + 1
	}

	return offset
}

func toggleSetting(idx int, m *tui_context.Model) bool {
	cfg := &m.Config
	reload := false
	switch idx {
	case 0:
		cfg.ShowHidden = !cfg.ShowHidden
		reload = true
	case 1:
		cfg.CaseSensitive = !cfg.CaseSensitive
		reload = true
	case 2:
		cfg.ConfirmOperations = !cfg.ConfirmOperations
	case 3:
		cfg.WrapNavigation = !cfg.WrapNavigation
	case 4:
		cfg.EditorIndex = (cfg.EditorIndex + 1) % len(constants.Editors)
	case 5:
		cfg.UseTrash = !cfg.UseTrash
	case 6:
		cfg.ShowHeader = !cfg.ShowHeader
		m.SyncViewportHeight()
	case 7:
		cfg.EnableGit = !cfg.EnableGit
		reload = true
	case 8:
		cfg.ShowSize = !cfg.ShowSize
		m.SyncViewportHeight()
		reload = true
	case 9:
		if cfg.ShowSize {
			cfg.SizeFormatIndex = (cfg.SizeFormatIndex + 1) % len(format.SizeFormats)
			reload = true
		}
	case 10:
		cfg.ShowDateModified = !cfg.ShowDateModified
		m.SyncViewportHeight()
		reload = true
	case 11:
		if cfg.ShowDateModified {
			cfg.DateFormatIndex = (cfg.DateFormatIndex + 1) % len(format.DateFormats)
			reload = true
		}
	case 12:
		cfg.ThemeIndex = (cfg.ThemeIndex + 1) % len(theme.Themes)
		// Update cached styles
		m.Display.Styles = theme.GetStylesheet(cfg.ThemeIndex)
		// Update spinner style for the new theme
		m.Display.LoadingSpinner.Style = m.Display.LoadingSpinner.Style.Foreground(theme.Themes[cfg.ThemeIndex].Dir)
	}
	_ = cfg.Save()
	return reload
}

func toggleSettingPrev(idx int, m *tui_context.Model) bool {
	cfg := &m.Config
	switch idx {
	case 4:
		cfg.EditorIndex = (cfg.EditorIndex - 1 + len(constants.Editors)) % len(constants.Editors)
	case 9:
		if cfg.ShowSize {
			cfg.SizeFormatIndex = (cfg.SizeFormatIndex - 1 + len(format.SizeFormats)) % len(format.SizeFormats)
			_ = cfg.Save()
			return true
		}
	case 11:
		if cfg.ShowDateModified {
			cfg.DateFormatIndex = (cfg.DateFormatIndex - 1 + len(format.DateFormats)) % len(format.DateFormats)
			_ = cfg.Save()
			return true
		}
	case 12:
		cfg.ThemeIndex = (cfg.ThemeIndex - 1 + len(theme.Themes)) % len(theme.Themes)
		// Update cached styles
		m.Display.Styles = theme.GetStylesheet(cfg.ThemeIndex)
		// Update spinner style for the new theme
		m.Display.LoadingSpinner.Style = m.Display.LoadingSpinner.Style.Foreground(theme.Themes[cfg.ThemeIndex].Dir)
	default:
		return toggleSetting(idx, m)
	}
	_ = cfg.Save()
	return false
}

// ConfirmSettingsReset resets all settings to defaults
func ConfirmSettingsReset(m *tui_context.Model) tea.Cmd {
	m.Config = config.DefaultConfig()
	_ = m.Config.Save()
	m.Display.Styles = theme.GetStylesheet(m.Config.ThemeIndex)
	m.UI.StopConfirming()
	m.Operations.ActionType = ""
	return tea.Batch(SetMsg(m, "Settings reset to defaults"), Reload(m, false))
}
