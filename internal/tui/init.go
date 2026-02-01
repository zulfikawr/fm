package tui

import (
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"
	"github.com/zulfikawr/fm/internal/tui/handlers/nav"
	"github.com/zulfikawr/fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
)

// Initialize sets up the initial commands for the TUI
func (a *App) Initialize() tea.Cmd {
	if a.Model.Config.EnableIcons {
		_ = theme.LoadIcons()
	}
	return tea.Batch(
		nav.Reload(a.Model, false),
		app.CheckForUpdates(),
		a.Model.Display.LoadingSpinner.Start(),
	)
}

// InitModel returns the initial command for the model
func InitModel(m *context.Model) tea.Cmd {
	return nav.Reload(m, false)
}

// Close releases resources held by the model
func Close(m *context.Model) {
	if m.Cancel != nil {
		m.Cancel()
	}
	if m.Watcher.Watcher != nil {
		_ = m.Watcher.Watcher.Close()
	}

	// Close all unique filesystems across all tabs
	closed := make(map[core.FileSystem]bool)
	if m.FS != nil {
		_ = m.FS.Close()
		closed[m.FS] = true
	}

	for _, t := range m.Tabs {
		if t.FS != nil && !closed[t.FS] {
			_ = t.FS.Close()
			closed[t.FS] = true
		}
	}
}
