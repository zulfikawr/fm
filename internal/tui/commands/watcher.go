package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// WatchDir returns a command that waits for a watcher event.
func WatchDir(watcher *fsnotify.Watcher) tea.Cmd {
	if watcher == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return WatcherClosedMsg{}
			}
			return WatchEventMsg{Event: event}
		case err, ok := <-watcher.Errors:
			if !ok {
				return WatcherClosedMsg{}
			}
			return WatcherErrorMsg{Err: err}
		}
	}
}
