package update

import (
	"fm/internal/config"
	"fm/internal/files/format"
	"fm/internal/files/ops"
	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"
	"fm/internal/tui/view"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleSettingsUpdate handles settings key events
func HandleSettingsUpdate(msg tea.Msg, m *state.Model) tea.Cmd {
	const numSettings = 29 // 13 settings + 16 keybindings

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", ".", "q":
			m.UI.SettingsOpen = false
			m.Config.Save()
			return actions.Reload(m)
		case "r":
			// Trigger reset confirmation
			m.UI.Confirming = true
			m.Operations.ActionType = "reset-settings"
			return nil
		case "y":
			// Confirm reset
			if m.UI.Confirming && m.Operations.ActionType == "reset-settings" {
				m.Config = config.DefaultConfig()
				m.Config.Save()
				m.UI.Confirming = false
				m.Operations.ActionType = ""
				return tea.Batch(commands.SetMsg(m, "Settings reset to defaults"), actions.Reload(m))
			}
		case "n":
			// Cancel reset
			if m.UI.Confirming && m.Operations.ActionType == "reset-settings" {
				m.UI.Confirming = false
				m.Operations.ActionType = ""
				return nil
			}
		case "up", "k":
			if m.UI.Confirming {
				return nil
			}
			if m.Settings.Cursor > 0 {
				m.Settings.Cursor--
				// Skip disabled settings
				if (m.Settings.Cursor == 9 && !m.Config.ShowSize) || (m.Settings.Cursor == 11 && !m.Config.ShowDateModified) {
					if m.Settings.Cursor > 0 {
						m.Settings.Cursor--
					}
				}
				// Update scroll offset
				view.UpdateSettingsScroll(m)
			}
		case "down", "j":
			if m.UI.Confirming {
				return nil
			}
			if m.Settings.Cursor < numSettings-1 {
				m.Settings.Cursor++
				// Skip disabled settings
				if (m.Settings.Cursor == 9 && !m.Config.ShowSize) || (m.Settings.Cursor == 11 && !m.Config.ShowDateModified) {
					if m.Settings.Cursor < numSettings-1 {
						m.Settings.Cursor++
					}
				}
				// Update scroll offset
				view.UpdateSettingsScroll(m)
			}
		case "enter", "right", "l", " ":
			if m.UI.Confirming {
				return nil
			}
			ToggleSetting(m.Settings.Cursor, m)
			UpdateItemFormatting(m)
			m.Config.Save()
		case "left", "h":
			if m.UI.Confirming {
				return nil
			}
			ToggleSettingPrev(m.Settings.Cursor, m)
			UpdateItemFormatting(m)
			m.Config.Save()
		}
	}
	return nil
}

// UpdateItemFormatting re-calculates all formatted strings for current items
func UpdateItemFormatting(m *state.Model) {
	layout := format.DateFormats[m.Config.DateFormatIndex].Layout
	for i := range m.Navigation.Items {
		item := &m.Navigation.Items[i]
		if !item.IsUp {
			item.FormattedSize = format.FormatSize(item.Size, m.Config.SizeFormatIndex)
			item.FormattedDate = item.MTime.Format(layout)
		}
	}
}

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
