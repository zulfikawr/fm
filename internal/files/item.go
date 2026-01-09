package files

import (
	"os"
	"time"
)

// Item represents a single file or directory entry in the file manager.
type Item struct {
	Name      string
	Path      string
	IsDir     bool
	Selected  bool   // Selected tracks if the item is marked for bulk operations.
	GitStatus string // GitStatus represents the porcelain status (e.g., "M", "A", "D").
	IsGhost   bool   // IsGhost indicates a file that exists in Git but not on disk (Deleted).
	Size      int64
	Mode      os.FileMode
	MTime     time.Time
	IsUp      bool // IsUp indicates if this item represents the parent directory ("..").
}
