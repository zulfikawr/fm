package state

import "fm/internal/files/core"

// ClipboardState holds clipboard data
type ClipboardState struct {
	Paths    []string
	SourceFS core.FileSystem
	IsCut    bool
	Action   string // "copy", "cut", "paste"
}

// Clear clears the clipboard
func (c *ClipboardState) Clear() {
	c.Paths = nil
	c.SourceFS = nil
	c.IsCut = false
	c.Action = ""
}

// SetCopy sets the clipboard for copy operation
func (c *ClipboardState) SetCopy(fs core.FileSystem, paths []string) {
	c.Paths = paths
	c.SourceFS = fs
	c.IsCut = false
	c.Action = "copy"
}

// SetCut sets the clipboard for cut operation
func (c *ClipboardState) SetCut(fs core.FileSystem, paths []string) {
	c.Paths = paths
	c.SourceFS = fs
	c.IsCut = true
	c.Action = "cut"
}
