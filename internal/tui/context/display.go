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

// DisplayState holds display and UI configuration state
type DisplayState struct {
	Width                int              // Terminal width
	Height               int              // Terminal height
	ViewportHeight       int              // Available height for file list
	SortMode             sorting.SortMode // Current sort mode
	LoadingSpinner       ui.Spinner       // Theme-aware loading spinner
	ReadOnly             bool             // True if current directory is on a read-only filesystem
	Styles               theme.Stylesheet // Cached stylesheet
	Layout               Layout           // Cached layout dimensions
	LastClickTime        time.Time        // For double-click detection
	LastClickIdx         int              // For double-click detection
	IsDragging           bool             // True if mouse is currently being dragged
	DragStartX           int              // Start X coordinate of drag
	DragStartY           int              // Start Y coordinate of drag
	DragEndX             int              // Current/End X coordinate of drag
	DragEndY             int              // Current/End Y coordinate of drag
	DragStartIdx         int              // Item index where drag started (-1 if none)
	InitialSelectedPaths map[string]bool  // Selection state before drag/shift start
}
