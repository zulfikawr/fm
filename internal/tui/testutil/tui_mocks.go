package testutil

import (
	"fm/internal/config"
	"fm/internal/constants"
	"fm/internal/files/sorting"
	"fm/internal/testutil"
	"fm/internal/tui/cache"
	"fm/internal/tui/state"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/fsnotify/fsnotify"
)

// CreateTestModel creates a model populated with mocks for testing
func CreateTestModel() *state.Model {
	watcher, _ := fsnotify.NewWatcher()
	mockFS := testutil.NewMockFileSystem()
	return &state.Model{
		FS: mockFS,
		GS: testutil.NewMockGitService(),
		Tabs: []state.Tab{{
			FS:            mockFS,
			Path:          "/test",
			SortMode:      sorting.SortDefault,
			SelectedPaths: make(map[string]bool),
		}},
		ActiveTab: 0,
		Navigation: state.NavigationState{
			Path:          "/test",
			SelectedPaths: make(map[string]bool),
		},
		Display: state.DisplayState{
			SortMode: sorting.SortDefault,
		},
		Inputs: state.InputState{
			ActiveInput: textinput.New(),
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
		Config: config.Config{},
	}
}
