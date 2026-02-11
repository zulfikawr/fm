package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	fileerrors "github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/logger"
	tuictx "github.com/zulfikawr/fm/internal/tui/context"
	tuierrors "github.com/zulfikawr/fm/internal/tui/errors"
	"github.com/zulfikawr/fm/internal/tui/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// SetMsg sets a temporary message in the footer
func SetMsg(m *tuictx.Model, msg string) tea.Cmd {
	m.Message.Push(msg, false)
	return tea.Tick(constants.MessageDisplayDuration, func(time.Time) tea.Msg {
		return messages.ClearMsg{}
	})
}

// SetErrMsg sets a temporary error message in the footer
func SetErrMsg(m *tuictx.Model, msg string) tea.Cmd {
	m.Message.Push(msg, true)
	return tea.Tick(constants.MessageDisplayDuration, func(time.Time) tea.Msg {
		return messages.ClearMsg{}
	})
}

// LogPush adds a new log entry
func LogPush(m *tuictx.Model, entry tuictx.LogEntry) string {
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	m.Logs.AddEntry(entry)
	return entry.ID
}

// LogUpdate updates an existing log entry
func LogUpdate(m *tuictx.Model, id string, entry tuictx.LogEntry) {
	m.Logs.UpdateStatus(id, entry)
}

// LogError logs a TUI error and sets it as the current error in the model
func LogError(m *tuictx.Model, err error, context string) tea.Cmd {
	if err == nil {
		return nil
	}

	var tuiErr *tuierrors.Error
	var ok bool
	if tuiErr, ok = err.(*tuierrors.Error); !ok {
		var fe *fileerrors.FileError
		if errors.As(err, &fe) {
			tuiErr = tuierrors.UserError(context, fe.Error())
		} else {
			tuiErr = tuierrors.SystemError(context, err)
		}
	}

	logger.Error(tuiErr.LogMessage())

	m.Message.Error = tuiErr
	return SetErrMsg(m, "Error: "+tuiErr.UserMessage())
}

// WatchDir returns a command that waits for a watcher event.
func WatchDir(watcher *fsnotify.Watcher) tea.Cmd {
	if watcher == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return messages.WatcherClosedMsg{}
			}
			return messages.WatchEventMsg{Event: event}
		case err, ok := <-watcher.Errors:
			if !ok {
				return messages.WatcherClosedMsg{}
			}
			return messages.WatcherErrorMsg{Err: err}
		}
	}
}

// WatchRemoteDir returns a command that waits for a polling interval.
func WatchRemoteDir() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return messages.RemotePollMsg{}
	})
}

// RestartWatcherAction restarts the file system watcher
func RestartWatcherAction(m *tuictx.Model) tea.Cmd {
	return func() tea.Msg {
		if m.Watcher.Watcher != nil {
			logger.CloseAndLog(m.Watcher.Watcher, "local filesystem watcher during restart")
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return messages.WatcherErrorMsg{Err: err}
		}

		m.Watcher.Watcher = watcher
		m.Watcher.LastWatched = ""
		m.Watcher.IsRemote = !m.FS.IsLocal()
		m.Watcher.IsListening = false

		return nil
	}
}
