package context

import (
	"time"

	"github.com/zulfikawr/fm/internal/files/sorting"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// --- Display State ---

// Layout holds the calculated dimensions for UI areas
type Layout struct {
	Width        int
	Height       int
	HeaderHeight int
	FooterHeight int
	BodyHeight   int
}

// Position represents a point in the terminal
type Position struct {
	X, Y int
}

// MouseState holds state for mouse interactions
type MouseState struct {
	LastClickTime    time.Time       // For double-click detection
	LastClickIdx     int             // For double-click detection
	IsDragging       bool            // True if mouse is currently being dragged
	DragStart        Position        // Start coordinate of drag
	DragEnd          Position        // Current/End coordinate of drag
	DragStartIdx     int             // Item index where drag started (-1 if none)
	InitialSelection map[string]bool // Selection state before drag/shift start
}

// DisplayState holds display and UI configuration state
type DisplayState struct {
	Width          int              // Terminal width
	Height         int              // Terminal height
	ViewportHeight int              // Available height for file list
	SortMode       sorting.SortMode // Current sort mode
	LoadingSpinner ui.Spinner       // Theme-aware loading spinner
	ReadOnly       bool             // True if current directory is on a read-only filesystem
	Styles         theme.Stylesheet // Cached stylesheet
	Layout         Layout           // Cached layout dimensions
	Mouse          MouseState       // Mouse interaction state
	RAMUsageMB     uint64           // Current RAM usage in MB
}
