package context

import (
	"context"
	"time"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/sorting"
	"github.com/zulfikawr/fm/internal/git"
	"github.com/zulfikawr/fm/internal/ssh"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/theme"

	"github.com/fsnotify/fsnotify"
)

// Model holds the application state.
type Model struct {
	// Core services
	FS     core.FileSystem
	GS     git.GitService
	Config config.Config

	// Lifecycle management
	Context context.Context
	Cancel  context.CancelFunc

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
	Search     SearchState     // Fuzzy content search state
	Logs       LogState        // Operation logs
}

// NewModel creates a new model with initial state
func NewModel(fs core.FileSystem, startPath string) *Model {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	styles := theme.GetStylesheet(cfg.ThemeIndex)

	ti := ui.NewInput(styles)
	ti.CharLimit = 256
	ti.Width = 30

	watcher, _ := fsnotify.NewWatcher()

	s := ui.NewSpinner(styles)

	m := &Model{
		FS:      fs,
		GS:      git.NewGitService(true),
		Config:  cfg,
		Context: ctx,
		Cancel:  cancel,
		Navigation: NavigationState{
			Path:          startPath,
			SelectedPaths: make(map[string]bool),
			LastShiftIdx:  -1,
		},
		Display: DisplayState{
			SortMode:       sorting.SortDefault,
			LoadingSpinner: s,
			Styles:         styles,
		},
		UI: UIState{
			PromptCache: make(map[string]string),
		},
		Inputs: InputState{
			ActiveInput: ti,
			Mode:        InputNone,
		},
		Cache: CacheState{
			CursorMemory:   core.NewSimpleCache[string, int](constants.MaxCacheEntries, 0), // No TTL for cursors
			OffsetMemory:   core.NewSimpleCache[string, int](constants.MaxCacheEntries, 0),
			GitStatusCache: core.NewSimpleCache[string, map[string]string](constants.MaxCacheEntries, 30*time.Second),
			GitRootCache:   core.NewSimpleCache[string, string](constants.MaxCacheEntries, 1*time.Hour),
			ItemCache:      core.NewSimpleCache[string, []core.Item](constants.MaxCacheEntries, 5*time.Minute),
		},
		Watcher: WatcherState{
			Watcher: watcher,
		},
		Operations: OperationsState{
			ProcessingItems: make(map[string]bool),
		},
		Remote: RemoteState{
			HostConfirmChan: make(chan *ssh.HostConfirmRequest),
		},
	}

	m.Tabs = []Tab{NewTab(TabOptions{
		FS:       fs,
		Path:     startPath,
		SortMode: sorting.SortDefault,
	})}
	m.ActiveTab = 0

	return m
}

// ClearSelection clears selection state across navigation and UI
func (m *Model) ClearSelection() {
	m.Navigation.ClearSelection()
	m.UI.SelectMode = false
}

// AddTab creates and appends a new tab to the model
func (m *Model) AddTab(path string) {
	if len(m.Tabs) >= 9 {
		return
	}
	m.Tabs = append(m.Tabs, NewTab(TabOptions{
		FS:         m.FS,
		Path:       path,
		SortMode:   m.Display.SortMode,
		RemoteUser: m.Remote.User,
		RemoteHost: m.Remote.Host,
	}))
}

// CloseActiveTab removes the current tab and adjusts the active index
func (m *Model) CloseActiveTab() bool {
	if len(m.Tabs) <= 1 {
		return false
	}
	if m.ActiveTab < 0 || m.ActiveTab >= len(m.Tabs) {
		m.ActiveTab = 0
		return false
	}

	m.Tabs = append(m.Tabs[:m.ActiveTab], m.Tabs[m.ActiveTab+1:]...)
	if m.ActiveTab >= len(m.Tabs) {
		m.ActiveTab = len(m.Tabs) - 1
	}
	return true
}

// SwitchTab switches to the specified tab number (1-based)
func (m *Model) SwitchTab(tabNum int) bool {
	if tabNum > 0 && tabNum <= len(m.Tabs) {
		m.ActiveTab = tabNum - 1
		return true
	}
	return false
}

// StartInput prepares the UI for a text input mode.
func (m *Model) StartInput(mode InputMode) {
	m.UI.InputActive = true
	m.UI.StartInput()
	m.Inputs.Mode = mode
	m.Inputs.ActiveInput.Reset()
}

// StopInput exits any active input mode and cleans up the input state.
func (m *Model) StopInput(clearInput bool) {
	m.UI.StopInput()
	m.Inputs.Mode = InputNone
	if clearInput {
		m.Inputs.ActiveInput.Reset()
	}
	m.Inputs.ActiveInput.Blur()
}

// SyncViewportHeight recalculates the available height for content areas.
func (m *Model) SyncViewportHeight() {
	// App Header(1) + App Footer(1) = 2
	h := m.Display.Height - 2

	// If we are in the file list and showing the header, subtract its height
	if !m.UI.SettingsOpen && !m.UI.LogOpen && !m.UI.ClipboardOpen &&
		m.Inputs.Mode != InputFuzzySearch && m.Config.ShowHeader {
		h -= 3 // List Header
	}

	if h < 1 {
		h = 1
	}
	m.Display.ViewportHeight = h
}
