package commands

import (
	"time"

	"fm/internal/constants"
	"fm/internal/files/core"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// LoadedItemsMsg is sent when directory contents have been loaded asynchronously.
type LoadedItemsMsg struct {
	Generation  int
	Path        string
	Items       []core.Item
	GitStatuses map[string]string
	GitBranch   string
	GitRoot     string
	IsReadOnly  bool
	Err         error
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

// WatcherErrorMsg is sent when the watcher encounters an error.
type WatcherErrorMsg struct {
	Err error
}

// WatcherClosedMsg is sent when the watcher is closed unexpectedly.
type WatcherClosedMsg struct{}

// ErrorMsg is sent when an operation fails.
type ErrorMsg struct{ Err error }

// ClearMsg is sent to clear the current status message.
type ClearMsg struct{}

// ConflictMsg is sent when a destination file already exists.
type ConflictMsg struct {
	Src          string
	Dst          string
	PendingItems []string
	IsMove       bool
}

// OperationFinishedMsg is sent when a background file operation completes.
type OperationFinishedMsg struct {
	Paths []string
}

// ProgressMsg is sent to update the progress bar.
type ProgressMsg struct {
	Percent float64
	Label   string
	Channel chan core.Progress
}

// SetMsg sets a temporary message in the footer and returns a command to clear it
func SetMsg(m *state.Model, msg string) tea.Cmd {
	m.Message.Text = msg
	m.Message.Time = time.Now()

	// Return command to clear message after delay
	return tea.Tick(constants.MessageDisplayDuration, func(time.Time) tea.Msg {
		return ClearMsg{}
	})
}

// ClearMsg clears the current status message and error
func ClearMsgState(m *state.Model) {
	m.Message.Text = ""
	m.Message.Error = nil
}
