package tui

import (
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/logger"
	"github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/handlers/app"
	"github.com/zulfikawr/fm/internal/tui/handlers/nav"
	"github.com/zulfikawr/fm/internal/tui/messages"
	"github.com/zulfikawr/fm/internal/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
)

// Initialize sets up the initial commands for the TUI
func (a *App) Initialize() tea.Cmd {
	if a.Model.Config.UI.EnableIcons {
		logger.LogIfError(theme.LoadIcons(), "tui: failed to load icons")
	}
	cmds := []tea.Cmd{
		nav.Reload(a.Model, false),
		app.CheckForUpdates(),
		a.Model.Display.LoadingSpinner.Start(),
		app.StartRAMTicker(), // Start RAM usage ticker
	}
	if a.Model.StartInAnalyzeMode {
		cmds = append(cmds, func() tea.Msg { return messages.StartAnalyzeMsg{} })
	}
	return tea.Batch(cmds...)
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
		logger.CloseAndLog(m.Watcher.Watcher, "tui watcher during shutdown")
	}

	// Close all unique filesystems across all tabs
	closed := make(map[core.FileSystem]bool)
	if m.FS != nil {
		logger.CloseAndLog(m.FS, "tui main filesystem during shutdown")
		closed[m.FS] = true
	}

	for i := range m.Tabs {
		t := m.Tabs[i]
		if t.FS != nil && !closed[t.FS] {
			logger.CloseAndLog(t.FS, "tui tab filesystem during shutdown")
			closed[t.FS] = true
		}
	}
}
