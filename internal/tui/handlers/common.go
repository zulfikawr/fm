package handlers

import (
	"fmt"
	"time"

	"fm/internal/constants"
	"fm/internal/files/core"
	"fm/internal/files/remote"
	"fm/internal/logger"
	"fm/internal/sshutil"
	tui_context "fm/internal/tui/context"
	tuierrors "fm/internal/tui/errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// --- Shared Messages ---

type LoadedItemsMsg struct {
	Generation  int
	Path        string
	Items       []core.Item
	GitStatuses map[string]string
	GitBranch   string
	GitRoot     string
	IsReadOnly  bool
	Cached      bool
	Err         error
}

type PartialItemsMsg struct {
	Generation int
	Path       string
	Items      []core.Item
}

type ErrorMsg struct {
	Err   error
	LogID string
}

type ClearMsg struct{}

type WatchEventMsg struct {
	Event fsnotify.Event
}

type WatcherErrorMsg struct {
	Err error
}

type WatcherClosedMsg struct{}

type DebounceWatchMsg struct{}

type DebounceFilterMsg struct {
	Generation int
}

type RemotePollMsg struct{}

type ProgressMsg struct {
	Percent float64
	Label   string
	Channel chan core.Progress
}

type OperationFinishedMsg struct {
	Paths []string
	LogID string
}

type ConflictMsg struct {
	Src          string
	Dst          string
	PendingItems []string
	IsMove       bool
	OpType       string
	LogID        string
}

type GitStatusMsg struct {
	Path     string
	Statuses map[string]string
	Branch   string
}

type RemoteConnectMsg struct {
	FS   core.FileSystem
	Path string
	Err  error
}

type HostConfirmMsg struct {
	Request *sshutil.HostConfirmRequest
}

type SearchMsg struct {
	Query   string
	Results []core.FileResult
	Err     error
}

// --- Shared Command Factories ---

// SetMsg sets a temporary message in the footer
func SetMsg(m *tui_context.Model, msg string) tea.Cmd {
	m.Message.Push(msg, false)
	return tea.Tick(constants.MessageDisplayDuration, func(time.Time) tea.Msg {
		return ClearMsg{}
	})
}

// SetErrMsg sets a temporary error message in the footer
func SetErrMsg(m *tui_context.Model, msg string) tea.Cmd {
	m.Message.Push(msg, true)
	return tea.Tick(constants.MessageDisplayDuration, func(time.Time) tea.Msg {
		return ClearMsg{}
	})
}

// LogPush adds a new log entry
func LogPush(m *tui_context.Model, opType string, level tui_context.LogLevel, status tui_context.LogStatus, message, details string) string {
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	m.Logs.AddEntry(tui_context.LogEntry{
		ID:        id,
		Timestamp: time.Now(),
		Type:      opType,
		Level:     level,
		Status:    status,
		Message:   message,
		Details:   details,
	})
	return id
}

// LogUpdate updates an existing log entry
func LogUpdate(m *tui_context.Model, id string, status tui_context.LogStatus, level tui_context.LogLevel, message, details string) {
	m.Logs.UpdateStatus(id, status, level, message, details)
}

// LogError logs a TUI error and sets it as the current error in the model
func LogError(m *tui_context.Model, err error, context string) tea.Cmd {
	if err == nil {
		return nil
	}

	var tuiErr *tuierrors.Error
	var ok bool
	if tuiErr, ok = err.(*tuierrors.Error); !ok {
		tuiErr = tuierrors.SystemError(context, err)
	}

	// In the new architecture, we might want a global error handler like before
	// For now, let's just log it and set it in the model
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

// WatchRemoteDir returns a command that waits for a polling interval.
func WatchRemoteDir() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return RemotePollMsg{}
	})
}

// RestartWatcherAction restarts the file system watcher
func RestartWatcherAction(m *tui_context.Model) tea.Cmd {
	return func() tea.Msg {
		if m.Watcher.Watcher != nil {
			m.Watcher.Watcher.Close()
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return WatcherErrorMsg{Err: err}
		}

		m.Watcher.Watcher = watcher
		m.Watcher.LastWatched = ""
		m.Watcher.IsRemote = !m.FS.IsLocal()
		m.Watcher.IsListening = false

		return nil
	}
}

func connectRemote(address, user, password, keyPath string, askChan chan *sshutil.HostConfirmRequest) tea.Cmd {
	return func() tea.Msg {
		hkcb, err := sshutil.GetHostKeyCallback(askChan)
		if err != nil {
			return RemoteConnectMsg{Err: err}
		}

		fs, err := remote.NewSftpFS(address, user, password, keyPath, hkcb)
		if err != nil {
			return RemoteConnectMsg{Err: err}
		}

		home, _ := fs.GetHomeDir()
		return RemoteConnectMsg{FS: fs, Path: home}
	}
}

func listenForHostConfirmation(askChan chan *sshutil.HostConfirmRequest) tea.Cmd {
	return func() tea.Msg {
		req, ok := <-askChan
		if !ok {
			return nil
		}
		return HostConfirmMsg{Request: req}
	}
}
