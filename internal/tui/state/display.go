package state

import (
	"fm/internal/files/sorting"

	"github.com/charmbracelet/bubbles/spinner"
)

// DisplayState holds display and UI configuration state
type DisplayState struct {
	Width          int              // Terminal width
	Height         int              // Terminal height
	ViewportHeight int              // Available height for file list
	SortMode       sorting.SortMode // Current sort mode
	LoadingSpinner spinner.Model    // Theme-aware loading spinner
	ReadOnly       bool             // True if current directory is on a read-only filesystem
}
