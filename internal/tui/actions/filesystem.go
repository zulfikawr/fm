package actions

import (
	"fm/internal/tui/commands"
	tuierrors "fm/internal/tui/errors"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// Reload triggers an asynchronous reload of the current directory
func Reload(m *state.Model) tea.Cmd {
	m.UI.Loading = true
	return commands.Reload(m.FS, m.GS, m.Navigation.Path, m.Navigation.PathGen, m.Display.SortMode, m.Config.ShowHidden)
}

// WatchDir starts watching the current directory
func WatchDir(m *state.Model) tea.Cmd {
	if !m.FS.IsLocal() {
		return nil
	}
	return commands.WatchDir(m.Watcher.Watcher)
}

// RestartWatcher restarts the file system watcher
func RestartWatcher(m *state.Model) tea.Cmd {
	return func() tea.Msg {
		// Close old watcher if it exists
		if m.Watcher.Watcher != nil {
			m.Watcher.Watcher.Close()
		}

		// Create new watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return commands.ErrorMsg{Err: err}
		}

		m.Watcher.Watcher = watcher
		m.Watcher.LastWatched = ""

		// Re-watch current directory
		if m.FS.IsLocal() && m.Navigation.Path != "" {
			if err := m.Watcher.Watcher.Add(m.Navigation.Path); err != nil {
				return commands.ErrorMsg{Err: err}
			}
			m.Watcher.LastWatched = m.Navigation.Path
		}

		// Resume watching
		return commands.WatchDir(m.Watcher.Watcher)()
	}
}

// HandleWatcherError handles watcher error messages
func HandleWatcherError(m *state.Model, err error) tea.Cmd {
	// Wrap watcher error as transient since it's retryable
	tuiErr := tuierrors.TransientError("file watcher", err.Error(), 3).
		WithContext("path", m.Navigation.Path)

	return tea.Batch(
		LogError(m, tuiErr, "File watcher error"),
		commands.SetMsg(m, "Watcher error: restarting..."),
		RestartWatcher(m),
	)
}

// HandleWatcherClosed handles watcher closed messages
func HandleWatcherClosed(m *state.Model) tea.Cmd {
	return tea.Batch(
		commands.SetMsg(m, "Watcher closed: restarting..."),
		RestartWatcher(m),
	)
}
