package actions

import (
	"fm/internal/config"
	"fm/internal/files/format"
	"fm/internal/files/ops"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
)

// ToggleSetting toggles the setting at the given index
func ToggleSetting(idx int, m *state.Model) {
	cfg := &m.Config
	switch idx {
	case 0:
		cfg.ShowHidden = !cfg.ShowHidden
	case 1:
		cfg.CaseSensitive = !cfg.CaseSensitive
	case 2:
		cfg.ConfirmOperations = !cfg.ConfirmOperations
	case 3:
		cfg.WrapNavigation = !cfg.WrapNavigation
	case 4:
		cfg.EditorIndex = (cfg.EditorIndex + 1) % len(ops.Editors)
	case 5:
		cfg.UseTrash = !cfg.UseTrash
	case 6:
		cfg.ShowHeader = !cfg.ShowHeader
	case 7:
		cfg.EnableGit = !cfg.EnableGit
	case 8:
		cfg.ShowSize = !cfg.ShowSize
	case 9:
		if cfg.ShowSize {
			cfg.SizeFormatIndex = (cfg.SizeFormatIndex + 1) % len(format.SizeFormats)
		}
	case 10:
		cfg.ShowDateModified = !cfg.ShowDateModified
	case 11:
		if cfg.ShowDateModified {
			cfg.DateFormatIndex = (cfg.DateFormatIndex + 1) % len(format.DateFormats)
		}
	case 12:
		cfg.ThemeIndex = (cfg.ThemeIndex + 1) % len(theme.Themes)
	}
}

// ToggleSettingPrev toggles the setting at the given index in reverse
func ToggleSettingPrev(idx int, m *state.Model) {
	cfg := &m.Config
	switch idx {
	case 4:
		cfg.EditorIndex = (cfg.EditorIndex - 1 + len(ops.Editors)) % len(ops.Editors)
	case 9:
		if cfg.ShowSize {
			cfg.SizeFormatIndex = (cfg.SizeFormatIndex - 1 + len(format.SizeFormats)) % len(format.SizeFormats)
		}
	case 11:
		if cfg.ShowDateModified {
			cfg.DateFormatIndex = (cfg.DateFormatIndex - 1 + len(format.DateFormats)) % len(format.DateFormats)
		}
	case 12:
		cfg.ThemeIndex = (cfg.ThemeIndex - 1 + len(theme.Themes)) % len(theme.Themes)
	default:
		ToggleSetting(idx, m)
	}
}

// ConfirmSettingsReset resets all settings to defaults
func ConfirmSettingsReset(m *state.Model) tea.Cmd {
	m.Config = config.DefaultConfig()
	m.Config.Save()
	m.UI.Confirming = false
	m.Operations.ActionType = ""
	return tea.Batch(commands.SetMsg(m, "Settings reset to defaults"), Reload(m))
}

// CancelSettingsReset cancels the settings reset
func CancelSettingsReset(m *state.Model) {
	m.UI.Confirming = false
	m.Operations.ActionType = ""
}
