package context

import (
	"context"
	"time"

	"fm/internal/constants"
	"fm/internal/files/conflict"
	"fm/internal/files/core"
)

// --- Operations State ---

// OperationsState holds file operation and action state
type OperationsState struct {
	Progress        ProgressState        // Progress tracking for operations
	ProcessingItems map[string]bool      // Paths currently being operated on (copy/move/delete)
	Clipboard       ClipboardState       // Clipboard state (cut/copy)
	Conflict        ConflictState        // Conflict resolution state
	ConflictPolicy  conflict.Policy      // Current conflict handling policy
	ActionType      constants.ActionType // "delete", "paste", "reset-settings", "conflict"
	CancelFunc      context.CancelFunc   // Function to cancel current operation
}

// --- Clipboard State ---

// ClipboardState holds clipboard data
type ClipboardState struct {
	Paths    []string
	SourceFS core.FileSystem
	IsCut    bool
	Action   string // "copy", "cut", "paste"
	Cursor   int    // Navigation cursor
	Offset   int    // Viewport offset
}

// Clear clears the clipboard
func (cs *ClipboardState) Clear() {
	cs.Paths = nil
	cs.SourceFS = nil
	cs.IsCut = false
	cs.Action = ""
	cs.Cursor = 0
	cs.Offset = 0
}

// SetCopy sets the clipboard for copy operation
func (cs *ClipboardState) SetCopy(fs core.FileSystem, paths []string) {
	cs.Paths = paths
	cs.SourceFS = fs
	cs.IsCut = false
	cs.Action = "copy"
	cs.Cursor = 0
	cs.Offset = 0
}

// SetCut sets the clipboard for cut operation
func (cs *ClipboardState) SetCut(fs core.FileSystem, paths []string) {
	cs.Paths = paths
	cs.SourceFS = fs
	cs.IsCut = true
	cs.Action = "cut"
	cs.Cursor = 0
	cs.Offset = 0
}

// --- Progress State ---

// ProgressState holds progress bar state
type ProgressState struct {
	Visible            bool
	Percent            float64
	Label              string
	LastProgressUpdate time.Time
}

// Show shows the progress bar with a label
func (ps *ProgressState) Show(label string) {
	ps.Visible = true
	ps.Label = label
	ps.Percent = 0
}

// Hide hides the progress bar
func (ps *ProgressState) Hide() {
	ps.Visible = false
	ps.Percent = 0
	ps.Label = ""
}

// Update updates the progress percentage
func (ps *ProgressState) Update(percent float64) {
	ps.Percent = percent
}

// --- Log State ---

// LogLevel defines the severity of a log entry
type LogLevel int

const (
	LogInfo LogLevel = iota
	LogSuccess
	LogWarn
	LogError
)

// LogStatus defines the current state of an operation
type LogStatus int

const (
	StatusPending LogStatus = iota
	StatusRunning
	StatusSuccess
	StatusError
)

// LogEntry represents a single entry in the operation log
type LogEntry struct {
	ID        string
	Timestamp time.Time
	Type      string    // "Copy", "Move", "Delete", "Search", "System"
	Level     LogLevel  // Info, Success, Warn, Error
	Status    LogStatus // Pending, Running, Success, Error
	Message   string
	Details   string // Full path, error stack, or extra info
}

// LogState holds the history of operations
type LogState struct {
	Entries []LogEntry
	Cursor  int // Current scroll position in the log view
	Offset  int // Viewport offset
}

// AddEntry adds a new entry to the log state
func (ls *LogState) AddEntry(entry LogEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	// Limit history to last 200 entries
	if len(ls.Entries) >= 200 {
		ls.Entries = ls.Entries[1:]
	}
	ls.Entries = append(ls.Entries, entry)
}

// UpdateStatus updates the status and level of an existing entry by ID
func (ls *LogState) UpdateStatus(id string, status LogStatus, level LogLevel, message string, details string) {
	for i := range ls.Entries {
		if ls.Entries[i].ID == id {
			ls.Entries[i].Status = status
			ls.Entries[i].Level = level
			if message != "" {
				ls.Entries[i].Message = message
			}
			if details != "" {
				ls.Entries[i].Details = details
			}
			break
		}
	}
}

// --- Message State ---

// Message represents a single status message
type Message struct {
	Text  string
	Time  time.Time
	IsErr bool
}

// MessageState holds status message state
type MessageState struct {
	Text  string    // Current status message (for backward compatibility)
	Time  time.Time // Time when message was set
	Error error     // Last error (if any)
	Stack []Message // Queue of messages to show
}

// Push adds a new message to the stack
func (ms *MessageState) Push(text string, isErr bool) {
	msg := Message{
		Text:  text,
		Time:  time.Now(),
		IsErr: isErr,
	}
	ms.Stack = append(ms.Stack, msg)
	ms.Text = text // Maintain compatibility
	ms.Time = msg.Time
}

// Pop removes the oldest message
func (ms *MessageState) Pop() {
	if len(ms.Stack) > 0 {
		ms.Stack = ms.Stack[1:]
		if len(ms.Stack) > 0 {
			ms.Text = ms.Stack[0].Text
			ms.Time = ms.Stack[0].Time
		} else {
			ms.Text = ""
		}
	}
}

// --- Conflict State ---

// ConflictState holds conflict resolution state for file operations
type ConflictState struct {
	Source       string   // Source path of the conflicting file
	Destination  string   // Destination path where conflict occurred
	PendingItems []string // Remaining items to process after conflict resolution
	IsMove       bool     // True if this is a move operation, false for copy
	OpType       string   // "copy", "move", "zip", "unzip"
	ApplyToAll   bool     // True if the current choice should apply to all remaining conflicts
	LogID        string   // Log ID associated with the operation
}

// Clear resets the conflict state
func (cf *ConflictState) Clear() {
	cf.Source = ""
	cf.Destination = ""
	cf.PendingItems = nil
	cf.IsMove = false
	cf.OpType = ""
	cf.ApplyToAll = false
	cf.LogID = ""
}

// Set initializes the conflict state
func (cf *ConflictState) Set(src, dst string, pending []string, isMove bool, opType string, logID string) {
	cf.Source = src
	cf.Destination = dst
	cf.PendingItems = pending
	cf.IsMove = isMove
	cf.OpType = opType
	cf.ApplyToAll = false
	cf.LogID = logID
}

// HasConflict returns true if there is an active conflict
func (cf *ConflictState) HasConflict() bool {
	return cf.Source != "" && cf.Destination != ""
}
