package tui

import (
	"fm/internal/tui/context"
	"fm/internal/tui/handlers"

	tea "github.com/charmbracelet/bubbletea"
)

// Initialize sets up the initial commands for the TUI
func (a *App) Initialize() tea.Cmd {
	return tea.Batch(
		handlers.Reload(a.Model, false),
		a.Model.Display.LoadingSpinner.Start(),
	)
}

// InitModel returns the initial command for the model
func InitModel(m *context.Model) tea.Cmd {
	return handlers.Reload(m, false)
}

// Close releases resources held by the model
func Close(m *context.Model) {
	if m.Cancel != nil {
		m.Cancel()
	}
	if m.Watcher.Watcher != nil {
		m.Watcher.Watcher.Close()
	}
	if m.FS != nil {
		m.FS.Close()
	}
}
