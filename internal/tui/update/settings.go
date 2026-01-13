package update

import (
	"fm/internal/tui/actions"
	"fm/internal/tui/state"
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
			// Update formatting once on exit
			for i := range m.Navigation.Items {
				m.Navigation.Items[i].UpdateFormatting(m.Config.SizeFormatIndex, m.Config.DateFormatIndex)
			}
			return actions.Reload(m)
		case "r":
			// Trigger reset confirmation
			m.UI.Confirming = true
			m.Operations.ActionType = "reset-settings"
			return nil
		case "y":
			// Confirm reset
			if m.UI.Confirming && m.Operations.ActionType == "reset-settings" {
				return actions.ConfirmSettingsReset(m)
			}
		case "n":
			// Cancel reset
			if m.UI.Confirming && m.Operations.ActionType == "reset-settings" {
				actions.CancelSettingsReset(m)
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
			actions.ToggleSetting(m.Settings.Cursor, m)
			m.Config.Save()
		case "left", "h":
			if m.UI.Confirming {
				return nil
			}
			actions.ToggleSettingPrev(m.Settings.Cursor, m)
			m.Config.Save()
		}
	}
	return nil
}
