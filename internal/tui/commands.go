package tui

import (
	"time"

	"filemanager/internal/files"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// LoadedItemsMsg is sent when directory contents have been loaded asynchronously.
type LoadedItemsMsg struct {
	Path        string
	Items       []files.Item
	GitStatuses map[string]string
	GitBranch   string
	Err         error
}

// WatchEventMsg is sent when a file system event occurs in the watched directory.
type WatchEventMsg struct {
	Event fsnotify.Event
	Err   error
}

type clearMsg struct{}

// reload triggers an asynchronous reload of the current directory.
func (m *Model) reload() tea.Cmd {
	m.loading = true
	path := m.path
	mode := m.sortMode
	showHidden := m.cfg.ShowHidden
	enableGit := m.cfg.EnableGit

	return func() tea.Msg {
		var gitStatuses map[string]string
		var gitBranch string
		if enableGit {
			gitStatuses, gitBranch = files.GetGitStatus(path)
		}

		items, err := files.Load(path, mode, showHidden, gitStatuses)
		return LoadedItemsMsg{
			Path:        path,
			Items:       items,
			GitStatuses: gitStatuses,
			GitBranch:   gitBranch,
			Err:         err,
		}
	}
}

func (m *Model) watchDir() tea.Cmd {
	return func() tea.Msg {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return nil
			}
			return WatchEventMsg{Event: event}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return nil
			}
			return WatchEventMsg{Err: err}
		}
	}
}

func (m *Model) setMsg(msg string) {
	m.msg = msg
	m.msgTime = time.Now()
}
