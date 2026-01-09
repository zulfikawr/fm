package tui

import (
	"time"

	"filemanager/internal/config"
	"filemanager/internal/files"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

// Model holds the application state.
type Model struct {
	path           string
	items          []files.Item // Original items
	filteredItems  []files.Item // Filtered items for display
	sortMode       files.SortMode
	cursorMemory   map[string]int // Path -> Cursor index
	offsetMemory   map[string]int // Path -> Scroll offset
	searchInput    textinput.Model
	renameInput    textinput.Model
	watcher        *fsnotify.Watcher
	lastWatched    string
	gitBranch      string
	searching      bool
	renaming       bool
	confirming     bool
	loading        bool
	settingsOpen   bool
	settingsCursor int // Index in the settings list
	selectMode     bool
	cfg            config.Config
	styles         Stylesheet
	actionType     string   // "delete" or "paste"
	clipboard      []string // Paths to copy
	cursor         int
	offset         int // Scroll offset
	width          int
	height         int
	msg            string
	msgTime        time.Time
	err            error
}

// NewModel creates and initializes a new Model starting in the specified path.
func NewModel(initialPath string) *Model {
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

	m := &Model{
		path:         initialPath,
		sortMode:     files.SortDefault,
		cursorMemory: make(map[string]int),
		offsetMemory: make(map[string]int),
		searchInput:  ti,
		renameInput:  ri,
		watcher:      watcher,
		cfg:          cfg,
		styles:       NewStylesheet(Themes[cfg.ThemeIndex]),
	}

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

// Init implements the tea.Model interface.
func (m *Model) Init() tea.Cmd {
	return m.reload()
}
