package messages

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
	Modified    int
	Staged      int
	Untracked   int
}

type PartialItemsMsg struct {
	Generation int
	Path       string
	Items      []core.Item
	GitRoot    string
	Branch     string
	Modified   int
	Staged     int
	Untracked  int
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
	Path      string
	Statuses  map[string]string
	Branch    string
	Modified  int
	Staged    int
	Untracked int
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

type ReloadMsg struct {
	Silent bool
}

type NavigateMsg struct {
	Path string
}

type RemoteGotoMsg struct {
	Input string
}

type StartCreateMsg struct{}

type StartConflictRenameMsg struct{}

type ResetSettingsMsg struct{}

type TabLimitMsg struct{}

type OpenFileMsg struct {
	Item core.Item
}

type ArchiveEnteredMsg struct {
	FS         core.FileSystem
	ParentFS   core.FileSystem
	ParentPath string
}

type PerformPasteMsg struct {
	OpName  string
	Message string
	Paths   []string
	DestDir string
	IsCut   bool
}

type PerformZipMsg struct {
	Targets []string
	Dst     string
	Message string
}

type PerformUnzipMsg struct {
	ZipPath string
	Dst     string
	Message string
}

type LogPushMsg struct {
	Type    string
	Message string
	Targets []string
}

type PerformRenameMsg struct {
	Selected core.Item
	OldPath  string
	NewPath  string
	NewName  string
}

type OperationFinishedEventMsg struct {
	LogID string
	Paths []string
}

type StatusMsg struct {
	Message string
	IsError bool
}

type ClearStatusMsg struct{}

type WatchDirMsg struct{}

type ReEnableMouseMsg struct{}

type CompletionsMsg struct {
	Completions []string
}

type IconsDownloadedMsg struct {
	Err error
}

type IconTestMsg struct {
	Success bool
}

type StartAnalyzeMsg struct{}

type AnalyzeFinishedMsg struct {
	Result *core.AnalysisResult
	Err    error
}
