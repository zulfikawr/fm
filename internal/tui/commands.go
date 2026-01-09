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

// DirSizeMsg is sent when a directory size has been calculated.
type DirSizeMsg struct {
	Path string
	Size int64
}

// GitStatusMsg is sent when git status has been fetched.
type GitStatusMsg struct {
	Path     string
	Statuses map[string]string
	Branch   string
}

// WatchEventMsg is sent when a file system event occurs in the watched directory.
type WatchEventMsg struct {
	Event fsnotify.Event
	Err   error
}

type errMsg struct{ err error }

type clearMsg struct{}

// reload triggers an asynchronous reload of the current directory.
func (m *Model) reload() tea.Cmd {
	m.loading = true
	path := m.path
	mode := m.sortMode
	showHidden := m.cfg.ShowHidden

	return func() tea.Msg {
		items, err := files.Load(path, mode, showHidden, nil)
		return LoadedItemsMsg{
			Path:        path,
			Items:       items,
			GitStatuses: nil,
			GitBranch:   "",
			Err:         err,
		}
	}
}

func calculateDirSize(path string) tea.Cmd {
	return func() tea.Msg {
		size := files.GetDirSize(path)
		return DirSizeMsg{Path: path, Size: size}
	}
}

func fetchGitStatus(path string) tea.Cmd {
	return func() tea.Msg {
		statuses, branch := files.GetGitStatus(path)
		return GitStatusMsg{Path: path, Statuses: statuses, Branch: branch}
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
