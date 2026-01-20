package handlers

import (
	"github.com/fsnotify/fsnotify"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/ssh"
)

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
	Request *ssh.HostConfirmRequest
}

type SearchMsg struct {
	Query   string
	Results []core.FileResult
	Err     error
}

type UpdateAvailableMsg struct {
	Version string
}

type UpdateFinishedMsg struct {
	Err error
}

type UpdateProgressMsg float64
