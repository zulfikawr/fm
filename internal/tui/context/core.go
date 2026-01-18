package context

import (
	"time"

	"fm/internal/files/core"
	"fm/internal/files/sorting"
	"fm/internal/tui/components/ui"
	"fm/internal/tui/theme"
)

// --- Navigation State ---

// NavigationState holds navigation-related state
type NavigationState struct {
	Path           string      // Current directory path
	PathGen        int         // Incremented on each navigation to detect stale messages
	Cursor         int         // Current cursor position
	Offset         int         // Scroll offset
	Items          []core.Item // Original items
	FilteredItems  []core.Item // Filtered items for display
	SelectedCount  int         // Number of selected items
	SelectedPaths  map[string]bool
	FilterTimer    *time.Timer     // Timer for debouncing filter
	FilterGen      int             // Generation counter for filter
	FilterQuery    string          // Current active filter query
	BackHistory    []string        // History for "Back" navigation
	ForwardHistory []string        // History for "Forward" navigation
	ParentFS       core.FileSystem // Previous FS before entering archive
	ParentPath     string          // Previous path before entering archive
}

// Select adds a path to the selection
func (n *NavigationState) Select(path string) {
	if n.SelectedPaths == nil {
		n.SelectedPaths = make(map[string]bool)
	}
	if !n.SelectedPaths[path] {
		n.SelectedPaths[path] = true
		n.SelectedCount++
	}
}

// Deselect removes a path from the selection
func (n *NavigationState) Deselect(path string) {
	if n.SelectedPaths != nil && n.SelectedPaths[path] {
		delete(n.SelectedPaths, path)
		n.SelectedCount--
	}
}

// ToggleSelection toggles the selection state of a path
func (n *NavigationState) ToggleSelection(path string) {
	if n.IsSelected(path) {
		n.Deselect(path)
	} else {
		n.Select(path)
	}
}

// SelectAll selects all currently filtered items
func (n *NavigationState) SelectAll() {
	if n.SelectedPaths == nil {
		n.SelectedPaths = make(map[string]bool)
	}
	for i := range n.FilteredItems {
		item := &n.FilteredItems[i]
		if item.IsUp {
			continue
		}
		if !n.SelectedPaths[item.Path] {
			n.SelectedPaths[item.Path] = true
			n.SelectedCount++
		}
		item.Selected = true
	}
}

// ClearSelection clears all selections
func (n *NavigationState) ClearSelection() {
	n.SelectedPaths = make(map[string]bool)
	n.SelectedCount = 0
	for i := range n.Items {
		n.Items[i].Selected = false
	}
	for i := range n.FilteredItems {
		n.FilteredItems[i].Selected = false
	}
}

// IsSelected checks if a path is selected
func (n *NavigationState) IsSelected(path string) bool {
	if n.SelectedPaths == nil {
		return false
	}
	return n.SelectedPaths[path]
}

// --- UI State ---

// UIState holds UI mode flags
type UIState struct {
	Confirming    bool
	SettingsOpen  bool
	LogOpen       bool
	ClipboardOpen bool
	Loading       bool
	SelectMode    bool
	InputActive   bool              // Consolidated flag for any text input (search, rename, etc)
	RemoteAuth    bool              // Specific flag for remote auth (uses input)
	HostConfirm   bool              // Waiting for known_hosts confirmation (uses y/n keys)
	PromptCache   map[string]string // Pre-calculated styled prompts
}

// Reset resets all UI flags to false
func (s *UIState) Reset() {
	s.Confirming = false
	s.SettingsOpen = false
	s.LogOpen = false
	s.ClipboardOpen = false
	s.Loading = false
	s.SelectMode = false
	s.InputActive = false
	s.RemoteAuth = false
	s.HostConfirm = false
	s.PromptCache = make(map[string]string)
}

// StartInput enters an input mode
func (s *UIState) StartInput() {
	s.InputActive = true
	s.LogOpen = false
	s.ClipboardOpen = false
	s.Confirming = false
}

// StopInput exits input mode
func (s *UIState) StopInput() {
	s.InputActive = false
}

// StartConfirming enters confirmation mode
func (s *UIState) StartConfirming() {
	s.Confirming = true
	s.InputActive = false
	s.LogOpen = false
	s.ClipboardOpen = false
}

// StopConfirming exits confirmation mode
func (s *UIState) StopConfirming() {
	s.Confirming = false
}

// ToggleSettings toggles the settings view
func (s *UIState) ToggleSettings() {
	s.SettingsOpen = !s.SettingsOpen
	if s.SettingsOpen {
		s.InputActive = false
		s.Confirming = false
		s.LogOpen = false
		s.ClipboardOpen = false
	}
}

// ToggleLogs toggles the log view
func (s *UIState) ToggleLogs() {
	s.LogOpen = !s.LogOpen
	if s.LogOpen {
		s.InputActive = false
		s.Confirming = false
		s.SettingsOpen = false
		s.ClipboardOpen = false
	}
}

// ToggleClipboard toggles the clipboard view
func (s *UIState) ToggleClipboard() {
	s.ClipboardOpen = !s.ClipboardOpen
	if s.ClipboardOpen {
		s.InputActive = false
		s.Confirming = false
		s.SettingsOpen = false
		s.LogOpen = false
	}
}

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
	Width          int              // Terminal width
	Height         int              // Terminal height
	ViewportHeight int              // Available height for file list
	SortMode       sorting.SortMode // Current sort mode
	LoadingSpinner ui.Spinner       // Theme-aware loading spinner
	ReadOnly       bool             // True if current directory is on a read-only filesystem
	Styles         theme.Stylesheet // Cached stylesheet
	Layout         Layout           // Cached layout dimensions
}

// --- Tab Management ---

// Tab represents a navigation context
type Tab struct {
	FS             core.FileSystem
	Path           string
	Items          []core.Item
	FilteredItems  []core.Item
	Cursor         int
	Offset         int
	SortMode       sorting.SortMode
	GitBranch      string
	GitRoot        string
	SearchQuery    string
	Searching      bool
	SelectMode     bool
	SelectedPaths  map[string]bool
	RemoteUser     string
	RemoteHost     string
	BackHistory    []string
	ForwardHistory []string
	ParentFS       core.FileSystem // Previous FS before entering archive
	ParentPath     string          // Previous path before entering archive
}

// NewTab creates a new tab for the given path
func NewTab(fs core.FileSystem, path string, sortMode sorting.SortMode, remoteUser, remoteHost string) Tab {
	return Tab{
		FS:            fs,
		Path:          path,
		SortMode:      sortMode,
		SelectedPaths: make(map[string]bool),
		RemoteUser:    remoteUser,
		RemoteHost:    remoteHost,
	}
}

// SelectedCount returns the number of selected items
func (t *Tab) SelectedCount() int {
	return len(t.SelectedPaths)
}

// IsSelected returns true if the given path is selected
func (t *Tab) IsSelected(path string) bool {
	if t.SelectedPaths == nil {
		return false
	}
	return t.SelectedPaths[path]
}

// --- Input State ---

// InputMode represents what the current active input is for.
type InputMode int

const (
	InputNone InputMode = iota
	InputSearch
	InputRename
	InputGoto
	InputAuth
	InputFuzzySearch
	InputZip
	InputUnzip
)

// InputState holds the unified text input model.
type InputState struct {
	ActiveInput ui.Input  // The single shared text input
	Mode        InputMode // What we are currently inputting
	AltMode     bool      // Toggled alternative mode (e.g., Remote for Goto, KeyPath for Auth)
}
