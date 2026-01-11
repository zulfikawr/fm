package tui

import (
	"time"

	"fm/internal/config"
	"fm/internal/files"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

// Model holds the application state.
type Model struct {
	fs              files.FileSystem
	tabs            []Tab // Multiple tabs
	activeTab       int   // Index of active tab
	path            string
	pathGeneration  int          // Incremented on each navigation to detect stale messages
	items           []files.Item // Original items
	filteredItems   []files.Item // Filtered items for display
	sortMode        files.SortMode
	cursorMemory    *SimpleCache                 // Path -> Cursor index (LRU cache)
	offsetMemory    *SimpleCache                 // Path -> Scroll offset (LRU cache)
	dirSizeCache    *LRUCache                    // Path -> Size (LRU cache)
	pendingDirSizes map[string]int64             // Batched sizes waiting to be applied
	gitStatusCache  map[string]map[string]string // Path -> Git status map (simple cache)
	searchInput     textinput.Model
	renameInput     textinput.Model
	watcher         *fsnotify.Watcher
	lastWatched     string
	lastWatchedGit  string
	gitBranch       string
	gitRoot         string
	searching       bool
	renaming        bool
	confirming      bool
	loading         bool
	sizeTickActive  bool
	showProgress    bool
	progressPercent float64
	progressLabel   string
	processingItems map[string]bool // Paths currently being operated on (copy/move/delete)
	selectedPaths   map[string]bool // Paths currently selected
	settingsOpen    bool
	settingsCursor  int // Index in the settings list
	settingsOffset  int // Scroll offset for settings
	selectMode      bool
	readOnly        bool // True if current directory is on a read-only filesystem
	cfg             config.Config
	styles          Stylesheet
	actionType      string   // "delete", "paste", "reset-settings", "conflict"
	clipboard       []string // Paths to copy/move
	clipboardCut    bool     // True if cut, false if copy
	conflictSrc     string   // Current source path in conflict
	conflictDst     string   // Current destination path in conflict
	pendingItems    []string // Items still waiting to be processed after a conflict
	cursor          int
	offset          int // Scroll offset
	width           int
	height          int
	msg             string
	msgTime         time.Time
	err             error
}

// NewModel creates and initializes a new Model starting in the specified path.
func NewModel(fs files.FileSystem, initialPath string) *Model {
	cfg := config.Load()

	ti := textinput.New()
	ti.Placeholder = "type to search"
	ti.Prompt = "/ "
	ti.CharLimit = 64
	ti.Width = 30

	ri := textinput.New()
	ri.Prompt = "New name: "
	ri.CharLimit = 64
	ri.Width = 30

	watcher, _ := fsnotify.NewWatcher()

	// Initialize with one tab
	initialTab := Tab{
		path:          initialPath,
		sortMode:      files.SortDefault,
		selectedPaths: make(map[string]bool),
	}

	m := &Model{
		fs:              fs,
		tabs:            []Tab{initialTab},
		activeTab:       0,
		path:            initialPath,
		sortMode:        files.SortDefault,
		cursorMemory:    NewSimpleCache(MaxCacheEntries),
		offsetMemory:    NewSimpleCache(MaxCacheEntries),
		dirSizeCache:    NewLRUCache(MaxCacheEntries),
		pendingDirSizes: make(map[string]int64),
		gitStatusCache:  make(map[string]map[string]string),
		searchInput:     ti,
		renameInput:     ri,
		watcher:         watcher,
		processingItems: make(map[string]bool),
		selectedPaths:   make(map[string]bool),
		cfg:             cfg,
		styles:          NewStylesheet(Themes[cfg.ThemeIndex]),
	}

	// Load persistent size cache
	m.dirSizeCache.Load(config.GetSizeCachePath())

	// Apply initial theme styles to inputs
	m.updateThemeStyles()

	return m
}

func (m *Model) updateThemeStyles() {
	theme := Themes[m.cfg.ThemeIndex]
	m.searchInput.TextStyle = m.searchInput.TextStyle.Background(theme.Bg)
	m.searchInput.PlaceholderStyle = m.searchInput.PlaceholderStyle.Background(theme.Bg).Foreground(lipgloss.Color("240"))
	m.searchInput.PromptStyle = m.searchInput.PromptStyle.Background(theme.Bg)
	m.renameInput.TextStyle = m.renameInput.TextStyle.Background(theme.Bg)
	m.renameInput.PromptStyle = m.renameInput.PromptStyle.Background(theme.Bg)
}

// Close releases resources held by the model, including filesystem connections and watchers.
func (m *Model) Close() {
	if m.watcher != nil {
		m.watcher.Close()
	}
	if m.fs != nil {
		m.fs.Close()
	}
}

// Init implements the tea.Model interface.
func (m *Model) Init() tea.Cmd {
	return m.reload()
}
