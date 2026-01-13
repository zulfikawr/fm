package actions

import (
	"fm/internal/tui/state"
	"fm/internal/tui/view"

	tea "github.com/charmbracelet/bubbletea"
)

// ResizeWindow handles window resize logic
func ResizeWindow(m *state.Model, msg tea.WindowSizeMsg) tea.Cmd {
	m.Display.Width = msg.Width
	m.Display.Height = msg.Height
	m.Display.ViewportHeight = view.CalculateViewportHeight(m)

	// Update input widths
	promptLen := len(m.Inputs.ActiveInput.Prompt)
	// Reserve space for prompt, margins (2), and a buffer (2)
	m.Inputs.ActiveInput.Width = msg.Width - promptLen - 4
	if m.Inputs.ActiveInput.Width < 10 {
		m.Inputs.ActiveInput.Width = 10
	}

	if m.UI.SettingsOpen {
		view.UpdateSettingsScroll(m)
	}

	return nil
}
