package state

// ConflictState holds conflict resolution state for file operations
type ConflictState struct {
	Source       string   // Source path of the conflicting file
	Destination  string   // Destination path where conflict occurred
	PendingItems []string // Remaining items to process after conflict resolution
	IsMove       bool     // True if this is a move operation, false for copy
}

// Clear resets the conflict state
func (c *ConflictState) Clear() {
	c.Source = ""
	c.Destination = ""
	c.PendingItems = nil
	c.IsMove = false
}

// Set initializes the conflict state
func (c *ConflictState) Set(src, dst string, pending []string, isMove bool) {
	c.Source = src
	c.Destination = dst
	c.PendingItems = pending
	c.IsMove = isMove
}

// HasConflict returns true if there is an active conflict
func (c *ConflictState) HasConflict() bool {
	return c.Source != "" && c.Destination != ""
}
