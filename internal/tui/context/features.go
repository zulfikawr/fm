package context

import (
	"context"

	"fm/internal/files/core"
	"fm/internal/ssh"
)

// --- Git State ---

// GitState holds git-related state
type GitState struct {
	Branch     string             // Current git branch name
	Root       string             // Git repository root path
	CancelFunc context.CancelFunc // Function to cancel the current git operation
}

// --- Remote State ---

// RemoteState holds remote connection state
type RemoteState struct {
	Host            string                       // For interactive remote connection
	User            string                       // For interactive remote connection
	HostConfirmChan chan *ssh.HostConfirmRequest // Channel for host confirmation requests
	HostConfirmReq  *ssh.HostConfirmRequest      // Current host confirmation request
}

// --- Search State ---

// SearchState holds the state for the fuzzy content search feature
type SearchState struct {
	Query       string             // The active search query
	Results     []core.FileResult  // Grouped search results
	IsSearching bool               // Whether a search is currently in progress
	CursorFile  int                // Index of the selected file
	CursorMatch int                // Index of the selected match within the selected file
	Offset      int                // Scroll offset
	CancelFunc  context.CancelFunc // Function to cancel the current search
}

// --- Settings State ---

// SettingsState holds settings view state
type SettingsState struct {
	Cursor int // Index in the settings list
	Offset int // Scroll offset for settings
}
