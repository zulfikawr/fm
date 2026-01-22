package app

import (
	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/logger"
	tui_context "github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/utils"
	"github.com/zulfikawr/fm/internal/tui/messages"
	"github.com/zulfikawr/fm/internal/tui/theme"

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
	totalItems := 38
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
		reload = ToggleSetting(m.Settings.Cursor, m)
	case "left", "h":
		reload = ToggleSettingPrev(m.Settings.Cursor, m)
	case "r":
		m.Operations.ActionType = constants.ActionResetSettings
		m.UI.StartConfirming()
	case "esc", "q":
		m.UI.ToggleSettings()
	}

	m.Settings.Offset = ScrollSettings(m)

	if reload {
		return func() tea.Msg { return messages.ReloadMsg{} }
	}
	return nil
}

// ScrollSettings recalculates the settings view offset
func ScrollSettings(m *tui_context.Model) int {
	cursor := m.Settings.Cursor
	offset := m.Settings.Offset
	height := m.Display.ViewportHeight

	rowIdx := 0
	if cursor <= 5 {
		rowIdx = cursor + 2
	} else if cursor <= 12 {
		rowIdx = cursor + 4
	} else if cursor <= 13 {
		rowIdx = cursor + 6
	} else {
		rowIdx = cursor + 8
	}

	if rowIdx < offset {
		newOffset := rowIdx
		if cursor == 0 || cursor == 6 || cursor == 13 || cursor == 14 {
			newOffset -= 2
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

func ToggleSetting(idx int, m *tui_context.Model) bool {
	cfg := m.Config // Copy
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
		cfg.EnableMouse = !cfg.EnableMouse
	case 13:
		cfg.ThemeIndex = (cfg.ThemeIndex + 1) % len(theme.Themes)
		m.Display.Styles = theme.GetStylesheet(cfg.ThemeIndex)
		m.Display.LoadingSpinner.Style = m.Display.LoadingSpinner.Style.Foreground(theme.Themes[cfg.ThemeIndex].Dir)
	}

	if err := cfg.Save(); err != nil {
		logger.Errorf("Failed to save config: %v", err)
	}
	m.Config = cfg // Explicit state change
	return reload
}

func ToggleSettingPrev(idx int, m *tui_context.Model) bool {
	cfg := m.Config // Copy
	var reload bool
	switch idx {
	case 4:
		cfg.EditorIndex = (cfg.EditorIndex - 1 + len(constants.Editors)) % len(constants.Editors)
	case 9:
		if cfg.ShowSize {
			cfg.SizeFormatIndex = (cfg.SizeFormatIndex - 1 + len(format.SizeFormats)) % len(format.SizeFormats)
			reload = true
		}
	case 11:
		if cfg.ShowDateModified {
			cfg.DateFormatIndex = (cfg.DateFormatIndex - 1 + len(format.DateFormats)) % len(format.DateFormats)
			reload = true
		}
	case 13:
		cfg.ThemeIndex = (cfg.ThemeIndex - 1 + len(theme.Themes)) % len(theme.Themes)
		m.Display.Styles = theme.GetStylesheet(cfg.ThemeIndex)
		m.Display.LoadingSpinner.Style = m.Display.LoadingSpinner.Style.Foreground(theme.Themes[cfg.ThemeIndex].Dir)
	default:
		return ToggleSetting(idx, m)
	}

	if err := cfg.Save(); err != nil {
		logger.Errorf("Failed to save config: %v", err)
	}
	m.Config = cfg // Explicit state change
	return reload
}

// ConfirmSettingsReset resets all settings to defaults
func ConfirmSettingsReset(m *tui_context.Model) tea.Cmd {
	newCfg := config.DefaultConfig()
	if err := newCfg.Save(); err != nil {
		logger.Errorf("Failed to save config: %v", err)
	}

	m.Config = newCfg
	m.Display.Styles = theme.GetStylesheet(m.Config.ThemeIndex)
	m.UI.StopConfirming()
	m.Operations.ActionType = ""

	return tea.Batch(utils.SetMsg(m, "Settings reset to defaults"), func() tea.Msg { return messages.ReloadMsg{} })
}
