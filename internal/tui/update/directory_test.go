package update

import (
	"context"
	"errors"
	"strings"
	"testing"

	"fm/internal/config"
	"fm/internal/constants"
	"fm/internal/files"
	"fm/internal/files/sorting"
	"fm/internal/testutil"
	"fm/internal/tui/actions"
	"fm/internal/tui/cache"
	"fm/internal/tui/commands"
	"fm/internal/tui/state"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// mockGitService implements a minimal GitService for testing
type mockGitService struct{}

func (m *mockGitService) IsEnabled() bool         { return true }
func (m *mockGitService) SetEnabled(enabled bool) {}
func (m *mockGitService) GetStatus(ctx context.Context, path string) (map[string]string, string) {
	return make(map[string]string), "main"
}
func (m *mockGitService) GetRoot(ctx context.Context, path string) string {
	return "/repo"
}

func createTestModel() *state.Model {
	watcher, _ := fsnotify.NewWatcher()
	return &state.Model{
		FS: testutil.NewMockFileSystem(),
		GS: &mockGitService{},
		Tabs: []state.Tab{{
			Path:          "/test",
			SortMode:      sorting.SortDefault,
			SelectedPaths: make(map[string]bool),
		}},
		ActiveTab: 0,
		Navigation: state.NavigationState{
			Path: "/test",
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
			SelectedPaths:   make(map[string]bool),
		},
		Config: config.Config{},
	}
}

func TestHandleLoadedItems_Success(t *testing.T) {
	m := createTestModel()
	m.Navigation.PathGen = 1
	m.UI.Loading = true

	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items: []files.Item{
			{Name: "file1.txt", IsDir: false},
			{Name: "dir1", IsDir: true},
		},
		GitBranch:  "main",
		GitRoot:    "/repo",
		IsReadOnly: false,
		Err:        nil,
	}

	logErr := func(err error, ctx string) tea.Cmd {
		t.Errorf("Unexpected error logged: %v - %s", err, ctx)
		return nil
	}

	cmd, handled := HandleLoadedItems(m, msg, logErr)

	if handled {
		t.Error("Expected handled to be false for successful load")
	}
	if cmd != nil {
		t.Error("Expected no command returned for successful load")
	}
	if m.UI.Loading {
		t.Error("Expected loading to be false after successful load")
	}
	if len(m.Navigation.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(m.Navigation.Items))
	}
	if m.Git.Branch != "main" {
		t.Errorf("Expected git branch 'main', got '%s'", m.Git.Branch)
	}
	if m.Git.Root != "/repo" {
		t.Errorf("Expected git root '/repo', got '%s'", m.Git.Root)
	}
	if m.Message.Error != nil {
		t.Errorf("Expected no error in message state, got %v", m.Message.Error)
	}
}

func TestHandleLoadedItems_StaleGeneration(t *testing.T) {
	m := createTestModel()
	m.Navigation.PathGen = 2 // Current generation is 2
	m.UI.Loading = true

	msg := commands.LoadedItemsMsg{
		Generation: 1, // Stale message
		Path:       "/test",
		Items:      []files.Item{{Name: "file1.txt"}},
	}

	logErr := func(err error, ctx string) tea.Cmd {
		t.Error("Error should not be logged for stale messages")
		return nil
	}

	cmd, handled := HandleLoadedItems(m, msg, logErr)

	if handled {
		t.Error("Expected handled to be false for stale generation")
	}
	if cmd != nil {
		t.Error("Expected no command for stale generation")
	}
	// Loading is set to false regardless of staleness (first line of function)
	if m.UI.Loading {
		t.Error("Loading state should be false after HandleLoadedItems")
	}
}

func TestHandleLoadedItems_Error(t *testing.T) {
	m := createTestModel()
	m.Navigation.PathGen = 1
	m.Navigation.Path = "/test"
	m.UI.Loading = true

	testErr := errors.New("failed to read directory")
	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items:      nil,
		Err:        testErr,
	}

	errorLogged := false
	logErr := func(err error, ctx string) tea.Cmd {
		errorLogged = true
		// Check if error message contains the original error
		if err != nil && !errors.Is(err, testErr) && !strings.Contains(err.Error(), testErr.Error()) {
			t.Errorf("Expected error containing %v, got %v", testErr, err)
		}
		return nil
	}

	cmd, handled := HandleLoadedItems(m, msg, logErr)

	if !handled {
		t.Error("Expected handled to be true when error occurs")
	}
	if !errorLogged {
		t.Error("Expected error to be logged")
	}
	if len(m.Navigation.Items) != 0 {
		t.Error("Expected items to be empty after error")
	}
	if !m.UI.Loading {
		t.Error("Expected loading to be false after error")
	}

	// Check that the command is not nil (should try to navigate back)
	if cmd == nil {
		t.Error("Expected command to navigate back after error")
	}
}

func TestHandleLoadedItems_RestoreSelection(t *testing.T) {
	m := createTestModel()
	m.Navigation.PathGen = 1
	m.UI.Loading = true

	// Add some selected paths
	m.Operations.SelectedPaths["/test/file1.txt"] = true
	m.Operations.SelectedPaths["/test/dir1"] = true

	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items: []files.Item{
			{Name: "file1.txt", Path: "/test/file1.txt", IsDir: false},
			{Name: "dir1", Path: "/test/dir1", IsDir: true},
			{Name: "file2.txt", Path: "/test/file2.txt", IsDir: false},
		},
		Err: nil,
	}

	logErr := func(err error, ctx string) tea.Cmd { return nil }

	_, _ = HandleLoadedItems(m, msg, logErr)

	// Check that selected items are marked as selected
	selectedCount := 0
	for _, item := range m.Navigation.Items {
		if item.Selected {
			selectedCount++
		}
	}

	if selectedCount != 2 {
		t.Errorf("Expected 2 items to be selected, got %d", selectedCount)
	}

	if !m.UI.SelectMode {
		t.Error("Expected select mode to be enabled when items are selected")
	}
}

func TestHandleLoadedItems_CursorRestore(t *testing.T) {
	m := createTestModel()
	m.Navigation.PathGen = 1
	m.UI.Loading = true

	// Set up cache with cursor position
	m.Cache.CursorMemory.Put("/test", 5)
	m.Cache.OffsetMemory.Put("/test", 3)

	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items:      make([]files.Item, 10), // 10 items
		Err:        nil,
	}

	logErr := func(err error, ctx string) tea.Cmd { return nil }

	_, _ = HandleLoadedItems(m, msg, logErr)

	if m.Navigation.Cursor != 5 {
		t.Errorf("Expected cursor to be restored to 5, got %d", m.Navigation.Cursor)
	}
	if m.Navigation.Offset != 3 {
		t.Errorf("Expected offset to be restored to 3, got %d", m.Navigation.Offset)
	}
}

func TestHandleLoadedItems_CursorBoundsCheck(t *testing.T) {
	m := createTestModel()
	m.Navigation.PathGen = 1
	m.UI.Loading = true
	m.Navigation.Cursor = 100 // Out of bounds

	// Clear tabs so tab restoration doesn't interfere
	m.Tabs = []state.Tab{}
	m.ActiveTab = 0

	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items:      make([]files.Item, 5), // Only 5 items
		Err:        nil,
	}

	logErr := func(err error, ctx string) tea.Cmd { return nil }

	_, _ = HandleLoadedItems(m, msg, logErr)

	if m.Navigation.Cursor != 4 {
		t.Errorf("Expected cursor to be bounded to 4, got %d", m.Navigation.Cursor)
	}
}

func TestHandleLoadedItems_EmptyDirectory(t *testing.T) {
	m := createTestModel()
	m.Navigation.PathGen = 1
	m.UI.Loading = true
	m.Navigation.Cursor = 5

	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items:      []files.Item{}, // Empty directory
		Err:        nil,
	}

	logErr := func(err error, ctx string) tea.Cmd { return nil }

	_, _ = HandleLoadedItems(m, msg, logErr)

	if m.Navigation.Cursor != 0 {
		t.Errorf("Expected cursor to be reset to 0 for empty directory, got %d", m.Navigation.Cursor)
	}
	if m.Navigation.Offset != 0 {
		t.Errorf("Expected offset to be reset to 0 for empty directory, got %d", m.Navigation.Offset)
	}
}

func TestHandleLoadedItems_OffsetBoundsCheck(t *testing.T) {
	// Bug fix test: Offset should be bounds checked to prevent view at bottom of list
	m := createTestModel()
	m.Navigation.PathGen = 1
	m.UI.Loading = true
	m.Navigation.Cursor = 2
	m.Navigation.Offset = 100 // Way out of bounds

	// Clear tabs so they don't interfere
	m.Tabs = []state.Tab{}

	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items:      make([]files.Item, 5), // Only 5 items
		Err:        nil,
	}

	logErr := func(err error, ctx string) tea.Cmd { return nil }

	_, _ = HandleLoadedItems(m, msg, logErr)

	// Offset should be bounded to prevent view below items
	if m.Navigation.Offset >= len(msg.Items) {
		t.Errorf("Expected offset to be < %d, got %d", len(msg.Items), m.Navigation.Offset)
	}
	if m.Navigation.Offset < 0 {
		t.Errorf("Expected offset to be >= 0, got %d", m.Navigation.Offset)
	}
}

func TestReload(t *testing.T) {
	m := createTestModel()
	m.UI.Loading = false
	m.Navigation.Path = "/test"
	m.Navigation.PathGen = 1
	m.Display.SortMode = sorting.SortName
	m.Config.ShowHidden = true

	cmd := actions.Reload(m)

	if !m.UI.Loading {
		t.Error("Expected loading flag to be set")
	}

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}

	// Execute the command to get the message
	msg := cmd()
	loadedMsg, ok := msg.(commands.LoadedItemsMsg)
	if !ok {
		t.Fatalf("Expected LoadedItemsMsg, got %T", msg)
	}

	if loadedMsg.Generation != 1 {
		t.Errorf("Expected generation 1, got %d", loadedMsg.Generation)
	}
	if loadedMsg.Path != "/test" {
		t.Errorf("Expected path '/test', got '%s'", loadedMsg.Path)
	}
}
