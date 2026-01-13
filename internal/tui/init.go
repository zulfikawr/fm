package tui

import (
	"fm/internal/config"
	"fm/internal/constants"
	"fm/internal/files/core"
	"fm/internal/files/sorting"
	"fm/internal/git"
	"fm/internal/sshutil"
	"fm/internal/tui/actions"
	"fm/internal/tui/cache"
	"fm/internal/tui/components/loading"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// NewModel creates and initializes a new Model starting in the specified path.
func NewModel(fs core.FileSystem, initialPath string) *state.Model {
	cfg := config.Load()

	// Initialize single unified input
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 30

	watcher, _ := fsnotify.NewWatcher()

	// Initialize with one tab
	initialTab := state.Tab{
		FS:            fs,
		Path:          initialPath,
		SortMode:      sorting.SortDefault,
		SelectedPaths: make(map[string]bool),
	}

	// Create theme-aware spinner
	loadingSpinner := loading.NewSpinner(theme.Themes[cfg.ThemeIndex])

	m := &state.Model{
		FS:        fs,
		GS:        git.NewGitService(true),
		Tabs:      []state.Tab{initialTab},
		ActiveTab: 0,
		Navigation: state.NavigationState{
			Path:          initialPath,
			SelectedPaths: make(map[string]bool),
		},
		Display: state.DisplayState{
			SortMode:       sorting.SortDefault,
			LoadingSpinner: loadingSpinner,
		},
		UI: state.UIState{
			PromptCache: make(map[string]string),
		},
		Inputs: state.InputState{
			ActiveInput: ti,
			Mode:        state.InputNone,
		},
		Cache: state.CacheState{
			CursorMemory:   cache.NewSimpleCache(constants.MaxCacheEntries),
			OffsetMemory:   cache.NewSimpleCache(constants.MaxCacheEntries),
			GitStatusCache: make(map[string]map[string]string),
		},
		Watcher: state.WatcherState{
			Watcher: watcher,
		},
		Operations: state.OperationsState{
			ProcessingItems: make(map[string]bool),
		},
		Remote: state.RemoteState{
			HostConfirmChan: make(chan *sshutil.HostConfirmRequest),
		},
		Config: cfg,
	}

	return m
}

// Close releases resources held by the model, including filesystem connections and watchers.
func Close(m *state.Model) {
	if m.Watcher.Watcher != nil {
		m.Watcher.Watcher.Close()
	}
	if m.FS != nil {
		m.FS.Close()
	}
}

// ModelInit implements the tea.Model interface.
func (a *App) ModelInit() tea.Cmd {
	return tea.Batch(
		actions.Reload(a.Model),
		a.Model.Display.LoadingSpinner.Tick,
	)
}
