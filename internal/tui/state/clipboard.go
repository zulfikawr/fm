package state

// ClipboardState holds clipboard data
type ClipboardState struct {
	Paths  []string
	IsCut  bool
	Action string // "copy", "cut", "paste"
}

// Clear clears the clipboard
func (c *ClipboardState) Clear() {
	c.Paths = nil
	c.IsCut = false
	c.Action = ""
}

// SetCopy sets the clipboard for copy operation
func (c *ClipboardState) SetCopy(paths []string) {
	c.Paths = paths
	c.IsCut = false
	c.Action = "copy"
}

// SetCut sets the clipboard for cut operation
func (c *ClipboardState) SetCut(paths []string) {
	c.Paths = paths
	c.IsCut = true
	c.Action = "cut"
}
