package constants

import "time"

// Application version
var AppVersion = "0.0.0-dev"

// Timeouts
const (
	DirectoryLoadTimeout    = 30 * time.Second
	FileOperationTimeout    = 5 * time.Minute
	GitCommandTimeout       = 10 * time.Second
	SSHConnectionTimeout    = 5 * time.Second
	MessageDisplayDuration  = 3 * time.Second
	ProgressDisplayDuration = 500 * time.Millisecond
)

// Limits
const (
	MaxDirectoryDepth = 50
	MaxCacheEntries   = 2000
	MaxTabs           = 9
	MaxFilenameLength = 255
	MaxSearchLength   = 64
	MaxCopyWorkers    = 16
	MaxReadDirWorkers = 32
	MaxSearchWorkers  = 8
	CopyBufferSize    = 1024 * 1024 // 1MB
)

// ActionType represents the type of pending action
type ActionType string

const (
	ActionNone          ActionType = ""
	ActionDelete        ActionType = "delete"
	ActionPaste         ActionType = "paste"
	ActionCopy          ActionType = "copy"
	ActionCut           ActionType = "cut"
	ActionConflict      ActionType = "conflict"
	ActionUpdate        ActionType = "update"
	ActionResetSettings ActionType = "reset-settings"
	ActionCancel        ActionType = "cancel"
	ActionGoto          ActionType = "goto"
	ActionAuth          ActionType = "auth"
	ActionTestIcons     ActionType = "test-icons"
)

// Editors lists supported text editors
var Editors = []string{
	"vim",
	"nvim",
	"nano",
	"emacs",
	"vi",
	"code",
}

// Data counts for validation
const (
	ThemeCount      = 9
	DateFormatCount = 4
	SizeFormatCount = 3
)
