package state

import (
	"context"
	"fm/internal/constants"
)

// OperationsState holds file operation and action state
type OperationsState struct {
	Progress        ProgressState        // Progress tracking for operations
	ProcessingItems map[string]bool      // Paths currently being operated on (copy/move/delete)
	Clipboard       ClipboardState       // Clipboard state (cut/copy)
	Conflict        ConflictState        // Conflict resolution state
	ActionType      constants.ActionType // "delete", "paste", "reset-settings", "conflict"
	CancelFunc      context.CancelFunc   // Function to cancel current operation
}
