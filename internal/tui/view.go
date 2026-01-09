package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the application UI.
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	var body string
	if m.settingsOpen {
		body = m.renderSettingsList(header, footer)
	} else {
		body = m.renderList(header, footer)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}
