package update

import (
	"strings"

	"fm/internal/tui/actions"
	"fm/internal/tui/commands"
	"fm/internal/tui/filter"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleFileSystemMsg delegates filesystem-related messages to specialized handlers
func HandleFileSystemMsg(m *state.Model, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case commands.LoadedItemsMsg:
		cmd, handled := HandleLoadedItems(m, msg)
		if handled {
			return cmd
		}
		filter.Apply(m)
		return tea.Batch(actions.WatchDir(m), commands.FetchGitStatus(m.GS, m.Navigation.Path))

	case commands.WatchEventMsg:
		return HandleWatchEvent(m, msg)

	case commands.WatcherErrorMsg:
		return actions.HandleWatcherError(m, msg.Err)

	case commands.WatcherClosedMsg:
		return actions.HandleWatcherClosed(m)
	}
	return nil
}

// HandleWatchEvent handles file system watch event messages
func HandleWatchEvent(m *state.Model, msg commands.WatchEventMsg) tea.Cmd {
	path := msg.Event.Name
	isGit := strings.Contains(path, "/.git/") || strings.HasSuffix(path, "/.git") ||
		strings.Contains(path, "\\.git\\") || strings.HasSuffix(path, "\\.git")

	if msg.Err != nil {
		return actions.WatchDir(m)
	}

	if isGit {
		// Invalidate git status cache for current path if git repo changed
		delete(m.Cache.GitStatusCache, m.Navigation.Path)
		// Fetch status again but don't reload full directory list to avoid loops
		return tea.Batch(commands.FetchGitStatus(m.GS, m.Navigation.Path), actions.WatchDir(m))
	}

	// Invalidate git status cache on regular file changes for accuracy
	delete(m.Cache.GitStatusCache, m.Navigation.Path)
	return actions.Reload(m)
}
