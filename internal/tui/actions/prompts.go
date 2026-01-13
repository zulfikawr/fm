package actions

import (
	"fm/internal/tui/state"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// OpenPrompt configures and opens the unified input field
func OpenPrompt(m *state.Model, mode state.InputMode, initialValue string) tea.Cmd {
	m.UI.StartInput()
	m.Inputs.Mode = mode
	m.Inputs.AltMode = false

	m.Inputs.ActiveInput.Focus()
	m.Inputs.ActiveInput.SetValue(initialValue)

	// Reset to default display mode
	m.Inputs.ActiveInput.EchoMode = textinput.EchoNormal

	switch mode {
	case state.InputSearch:
		m.Inputs.ActiveInput.Prompt = "/ "
		m.Inputs.ActiveInput.Placeholder = "type to search"
	case state.InputRename:
		m.Inputs.ActiveInput.Prompt = "New name: "
		m.Inputs.ActiveInput.Placeholder = ""
	case state.InputGoto:
		m.Inputs.ActiveInput.Prompt = "Go to: "
		m.Inputs.ActiveInput.Placeholder = "path or user@host"
	case state.InputAuth:
		m.Inputs.ActiveInput.Prompt = "Auth: "
		m.Inputs.ActiveInput.Placeholder = ""
		m.Inputs.ActiveInput.EchoMode = textinput.EchoPassword
	}

	// Update width based on prompt to prevent wrapping
	if m.Display.Width > 0 {
		promptLen := len(m.Inputs.ActiveInput.Prompt)
		// Reserve space for prompt, margins (2), and a buffer (2)
		availableWidth := m.Display.Width - promptLen - 4

		// Reserve extra space for hints if in Goto or Auth mode
		if mode == state.InputGoto {
			availableWidth -= 13 // Length of "[Tab] Remote "
		} else if mode == state.InputAuth {
			availableWidth -= 15 // Length of "[Tab] Key Path "
		}

		if availableWidth < 10 {
			availableWidth = 10
		}
		m.Inputs.ActiveInput.Width = availableWidth
	}

	return textinput.Blink
}

// ClosePrompt closes the unified input field
func ClosePrompt(m *state.Model) {
	m.UI.StopInput()
	m.Inputs.Mode = state.InputNone
	m.Inputs.ActiveInput.Blur()
	m.Inputs.ActiveInput.SetValue("")
}
