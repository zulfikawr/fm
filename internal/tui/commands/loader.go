package commands

import (
	"context"

	"fm/internal/constants"
	"fm/internal/files"
	"fm/internal/files/listing"
	"fm/internal/files/sorting"
	"fm/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

// Reload triggers an asynchronous reload of the directory.
func Reload(fs files.FileSystem, gs git.GitService, path string, generation int, mode sorting.SortMode, showHidden bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), constants.DirectoryLoadTimeout)
		defer cancel()

		var gitStatuses map[string]string
		var gitBranch string
		var gitRoot string

		if gs.IsEnabled() {
			gitStatuses, gitBranch = gs.GetStatus(ctx, path)
			gitRoot = gs.GetRoot(ctx, path)
		}

		items, err := listing.Load(ctx, fs, path, mode, showHidden, gitStatuses)
		ro, _ := fs.IsReadOnly(ctx, path)

		return LoadedItemsMsg{
			Generation:  generation,
			Path:        path,
			Items:       items,
			GitStatuses: gitStatuses,
			GitBranch:   gitBranch,
			GitRoot:     gitRoot,
			IsReadOnly:  ro,
			Err:         err,
		}
	}
}

// FetchGitStatus triggers an asynchronous fetch of git status.
func FetchGitStatus(gs git.GitService, path string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), constants.GitCommandTimeout)
		defer cancel()

		statuses, branch := gs.GetStatus(ctx, path)
		return GitStatusMsg{Path: path, Statuses: statuses, Branch: branch}
	}
}
