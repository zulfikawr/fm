package tui

import (
	"fmt"
	"time"

	"filemanager/internal/files"
	"filemanager/internal/logger"

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
	fs := m.fs

	return func() tea.Msg {
		items, err := files.Load(fs, path, mode, showHidden, nil)
		return LoadedItemsMsg{
			Path:        path,
			Items:       items,
			GitStatuses: nil,
			GitBranch:   "",
			Err:         err,
		}
	}
}

func calculateDirSize(fs files.FileSystem, path string) tea.Cmd {
	return func() tea.Msg {
		size := files.GetDirSize(fs, path)
		return DirSizeMsg{Path: path, Size: size}
	}
}

func fetchGitStatus(fs files.FileSystem, path string) tea.Cmd {
	return func() tea.Msg {
		statuses, branch := fs.GetGitStatus(path)
		return GitStatusMsg{Path: path, Statuses: statuses, Branch: branch}
	}
}

func (m *Model) watchDir() tea.Cmd {
	if !m.fs.IsLocal() {
		return nil
	}
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

func deleteItems(fs files.FileSystem, targets []string, useTrash bool) tea.Cmd {
	return func() tea.Msg {
		var lastErr error
		count := 0
		for _, t := range targets {
			var err error
			if useTrash {
				err = files.Trash(fs, t)
			} else {
				err = files.Delete(fs, t)
			}

			if err != nil {
				logger.Error(fmt.Sprintf("Error deleting %s: %v", t, err))
				lastErr = err
				continue
			}
			count++
		}
		if lastErr != nil && count == 0 {
			return errMsg{err: lastErr}
		}
		logger.Info(fmt.Sprintf("Successfully deleted %d items", count))
		return clearMsg{}
	}
}
func pasteItems(fs files.FileSystem, clipboard []string, destDir string) tea.Cmd {
	return func() tea.Msg {
		for _, src := range clipboard {
			dst := fs.Join(destDir, fs.Base(src))
			if err := files.Copy(fs, src, dst); err != nil {
				logger.Error(fmt.Sprintf("Error copying %s to %s: %v", src, dst, err))
				return errMsg{err: err}
			}
		}
		logger.Info(fmt.Sprintf("Successfully pasted %d items to %s", len(clipboard), destDir))
		return clearMsg{}
	}
}
