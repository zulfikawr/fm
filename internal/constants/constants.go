package constants

import "time"

// Timeouts
const (
	DirectoryLoadTimeout   = 30 * time.Second
	FileOperationTimeout   = 5 * time.Minute
	GitCommandTimeout      = 10 * time.Second
	SSHConnectionTimeout   = 5 * time.Second
	MessageDisplayDuration = 3 * time.Second
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
	ActionResetSettings ActionType = "reset-settings"
	ActionCancel        ActionType = "cancel"
)
