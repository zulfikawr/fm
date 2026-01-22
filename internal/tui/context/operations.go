package context

import (
	"context"
	"time"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
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

// ConflictParams encapsulates data for setting conflict state
type ConflictParams struct {
	Source       string
	Destination  string
	PendingItems []string
	IsMove       bool
	OpType       string
	LogID        string
}

// Set initializes the conflict state
func (cf *ConflictState) Set(params ConflictParams) {
	cf.Source = params.Source
	cf.Destination = params.Destination
	cf.PendingItems = params.PendingItems
	cf.IsMove = params.IsMove
	cf.OpType = params.OpType
	cf.ApplyToAll = false
	cf.LogID = params.LogID
}

// HasConflict returns true if there is an active conflict
func (cf *ConflictState) HasConflict() bool {
	return cf.Source != "" && cf.Destination != ""
}
