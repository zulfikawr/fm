package state

import (
	"fm/internal/config"
	"fm/internal/files/core"
	"fm/internal/git"
)

// Model holds the application state.
type Model struct {
	// Core services
	FS     core.FileSystem
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
