package state

import (
	"time"

	"fm/internal/config"
	"fm/internal/constants"
	"fm/internal/files"
	"fm/internal/files/sorting"
	"fm/internal/git"
	"fm/internal/sshutil"
	"fm/internal/tui/cache"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/fsnotify/fsnotify"
)

// NavigationState holds navigation-related state
type NavigationState struct {
	Path          string       // Current directory path
	PathGen       int          // Incremented on each navigation to detect stale messages
	Cursor        int          // Current cursor position
	Offset        int          // Scroll offset
	Items         []files.Item // Original items
	FilteredItems []files.Item // Filtered items for display
	SelectedCount int          // Number of selected items
}

// DisplayState holds display and UI configuration state
type DisplayState struct {
	Width          int              // Terminal width
	Height         int              // Terminal height
	ViewportHeight int              // Available height for file list
	SortMode       sorting.SortMode // Current sort mode
	LoadingSpinner spinner.Model    // Theme-aware loading spinner
	ReadOnly       bool             // True if current directory is on a read-only filesystem
}

// OperationsState holds file operation and action state
type OperationsState struct {
	Progress        ProgressState        // Progress tracking for operations
	ProcessingItems map[string]bool      // Paths currently being operated on (copy/move/delete)
	SelectedPaths   map[string]bool      // Paths currently selected
	Clipboard       ClipboardState       // Clipboard state (cut/copy)
	Conflict        ConflictState        // Conflict resolution state
	ActionType      constants.ActionType // "delete", "paste", "reset-settings", "conflict"
}

// InputMode represents what the current active input is for.
type InputMode int

const (
	InputNone InputMode = iota
	InputSearch
	InputRename
	InputGoto
	InputAuth
)

// InputState holds the unified text input model.
type InputState struct {
	ActiveInput textinput.Model // The single shared text input
	Mode        InputMode       // What we are currently inputting
}

// CacheState holds caching-related state
type CacheState struct {
	CursorMemory   *cache.SimpleCache           // Path -> Cursor index (LRU cache)
	OffsetMemory   *cache.SimpleCache           // Path -> Scroll offset (LRU cache)
	GitStatusCache map[string]map[string]string // Path -> Git status map (simple cache)
}

// WatcherState holds filesystem and git watching state
type WatcherState struct {
	Watcher        *fsnotify.Watcher // Filesystem watcher
	LastWatched    string            // Last watched directory path
	LastWatchedGit string            // Last watched git repository path
}

// GitState holds git-related state
type GitState struct {
	Branch string // Current git branch name
	Root   string // Git repository root path
}

// RemoteState holds remote connection state
type RemoteState struct {
	Host            string                           // For interactive remote connection
	User            string                           // For interactive remote connection
	HostConfirmChan chan *sshutil.HostConfirmRequest // Channel for host confirmation requests
	HostConfirmReq  *sshutil.HostConfirmRequest      // Current host confirmation request
}

// SettingsState holds settings view state
type SettingsState struct {
	Cursor int // Index in the settings list
	Offset int // Scroll offset for settings
}

// MessageState holds status message state
type MessageState struct {
	Text  string    // Current status message
	Time  time.Time // Time when message was set
	Error error     // Last error (if any)
}

// Model holds the application state.
type Model struct {
	// Core services
	FS     files.FileSystem
	GS     git.GitService
	Config config.Config

	// Tab management
	Tabs      []Tab // Multiple tabs
	ActiveTab int   // Index of active tab

	// Grouped state
	Navigation NavigationState // Navigation and items
	Display    DisplayState    // Display and UI configuration
	UI         UIState         // UI mode and state flags
	Operations OperationsState // File operations and actions
	Inputs     InputState      // Text input models
	Cache      CacheState      // Caching state
	Watcher    WatcherState    // Filesystem watching
	Git        GitState        // Git information
	Remote     RemoteState     // Remote connection state
	Settings   SettingsState   // Settings view state
	Message    MessageState    // Status messages
}
