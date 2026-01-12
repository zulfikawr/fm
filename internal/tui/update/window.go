package update

import (
	"fm/internal/tui/state"
	"fm/internal/tui/view"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleWindowMsg delegates window-related messages to specialized handlers
func HandleWindowMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return HandleWindowSize(m, msg)
	}
	return nil
}

// HandleWindowSize handles window resize events
func HandleWindowSize(m *state.Model, msg tea.WindowSizeMsg) tea.Cmd {
	m.Display.Width = msg.Width
	m.Display.Height = msg.Height
	m.Display.ViewportHeight = view.CalculateViewportHeight(m)

	// Update input widths
	m.Inputs.ActiveInput.Width = msg.Width - 10
	if m.Inputs.ActiveInput.Width < 20 {
		m.Inputs.ActiveInput.Width = 20
	}

	if m.UI.SettingsOpen {
		view.UpdateSettingsScroll(m)
	}

	return nil
}
