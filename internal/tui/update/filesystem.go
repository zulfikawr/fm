package update

import (
	"strings"

	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/filter"
	"fm/internal/tui/state"

	tuierrors "fm/internal/tui/errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// HandleFileSystemMsg delegates filesystem-related messages to specialized handlers
func HandleFileSystemMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case commands.LoadedItemsMsg:
		cmd, handled := HandleLoadedItems(m, msg, func(err error, context string) tea.Cmd {
			return actions.LogError(m, err, context)
		})
		if handled {
			return cmd
		}
		filter.Apply(m)
		return tea.Batch(WatchDir(m), commands.FetchGitStatus(m.GS, m.Navigation.Path))

	case commands.WatchEventMsg:
		return HandleWatchEvent(m, msg)

	case commands.WatcherErrorMsg:
		return HandleWatcherError(m, msg, func(err error, context string) tea.Cmd {
			return actions.LogError(m, err, context)
		}, func(s string) tea.Cmd { return commands.SetMsg(m, s) })

	case commands.WatcherClosedMsg:
		return HandleWatcherClosed(m, func(s string) tea.Cmd { return commands.SetMsg(m, s) })
	}
	return nil
}

// HandleWatchEvent handles file system watch event messages
func HandleWatchEvent(m *state.Model, msg commands.WatchEventMsg) tea.Cmd {
	path := msg.Event.Name
	isGit := strings.Contains(path, "/.git/") || strings.HasSuffix(path, "/.git") ||
		strings.Contains(path, "\\.git\\") || strings.HasSuffix(path, "\\.git")

	if msg.Err != nil {
		return WatchDir(m)
	}

	if isGit {
		// Invalidate git status cache for current path if git repo changed
		delete(m.Cache.GitStatusCache, m.Navigation.Path)
		// Fetch status again but don't reload full directory list to avoid loops
		return tea.Batch(commands.FetchGitStatus(m.GS, m.Navigation.Path), WatchDir(m))
	}

	// Invalidate git status cache on regular file changes for accuracy
	delete(m.Cache.GitStatusCache, m.Navigation.Path)
	return actions.Reload(m)
}

// HandleWatcherError handles watcher error messages
func HandleWatcherError(m *state.Model, msg commands.WatcherErrorMsg, logError func(error, string) tea.Cmd, setMsg func(string) tea.Cmd) tea.Cmd {
	// Wrap watcher error as transient since it's retryable
	err := tuierrors.TransientError("file watcher", msg.Err.Error(), 3).
		WithContext("path", m.Navigation.Path)

	return tea.Batch(logError(err, "File watcher error"), setMsg("Watcher error: restarting..."), RestartWatcher(m))
}

// HandleWatcherClosed handles watcher closed messages
func HandleWatcherClosed(m *state.Model, setMsg func(string) tea.Cmd) tea.Cmd {
	return tea.Batch(setMsg("Watcher closed: restarting..."), RestartWatcher(m))
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
