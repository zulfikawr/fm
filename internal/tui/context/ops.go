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
func (c *ClipboardState) Clear() {
	c.Paths = nil
	c.SourceFS = nil
	c.IsCut = false
	c.Action = ""
	c.Cursor = 0
	c.Offset = 0
}

// SetCopy sets the clipboard for copy operation
func (c *ClipboardState) SetCopy(fs core.FileSystem, paths []string) {
	c.Paths = paths
	c.SourceFS = fs
	c.IsCut = false
	c.Action = "copy"
	c.Cursor = 0
	c.Offset = 0
}

// SetCut sets the clipboard for cut operation
func (c *ClipboardState) SetCut(fs core.FileSystem, paths []string) {
	c.Paths = paths
	c.SourceFS = fs
	c.IsCut = true
	c.Action = "cut"
	c.Cursor = 0
	c.Offset = 0
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
func (p *ProgressState) Show(label string) {
	p.Visible = true
	p.Label = label
	p.Percent = 0
}

// Hide hides the progress bar
func (p *ProgressState) Hide() {
	p.Visible = false
	p.Percent = 0
	p.Label = ""
}

// Update updates the progress percentage
func (p *ProgressState) Update(percent float64) {
	p.Percent = percent
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
func (s *LogState) AddEntry(entry LogEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	// Limit history to last 200 entries
	if len(s.Entries) >= 200 {
		s.Entries = s.Entries[1:]
	}
	s.Entries = append(s.Entries, entry)
}

// UpdateStatus updates the status and level of an existing entry by ID
func (s *LogState) UpdateStatus(id string, status LogStatus, level LogLevel, message string, details string) {
	for i := range s.Entries {
		if s.Entries[i].ID == id {
			s.Entries[i].Status = status
			s.Entries[i].Level = level
			if message != "" {
				s.Entries[i].Message = message
			}
			if details != "" {
				s.Entries[i].Details = details
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
func (s *MessageState) Push(text string, isErr bool) {
	msg := Message{
		Text:  text,
		Time:  time.Now(),
		IsErr: isErr,
	}
	s.Stack = append(s.Stack, msg)
	s.Text = text // Maintain compatibility
	s.Time = msg.Time
}

// Pop removes the oldest message
func (s *MessageState) Pop() {
	if len(s.Stack) > 0 {
		s.Stack = s.Stack[1:]
		if len(s.Stack) > 0 {
			s.Text = s.Stack[0].Text
			s.Time = s.Stack[0].Time
		} else {
			s.Text = ""
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
func (c *ConflictState) Clear() {
	c.Source = ""
	c.Destination = ""
	c.PendingItems = nil
	c.IsMove = false
	c.OpType = ""
	c.ApplyToAll = false
	c.LogID = ""
}

// Set initializes the conflict state
func (c *ConflictState) Set(src, dst string, pending []string, isMove bool, opType string, logID string) {
	c.Source = src
	c.Destination = dst
	c.PendingItems = pending
	c.IsMove = isMove
	c.OpType = opType
	c.ApplyToAll = false
	c.LogID = logID
}

// HasConflict returns true if there is an active conflict
func (c *ConflictState) HasConflict() bool {
	return c.Source != "" && c.Destination != ""
}
