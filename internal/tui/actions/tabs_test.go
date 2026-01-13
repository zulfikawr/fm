package actions

import (
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/files/sorting"
	"fm/internal/tui/state"
)

func TestCreateTab_Success(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 0
	m.Navigation.Path = "/test2"
	m.Navigation.Cursor = 5
	m.Navigation.Offset = 3

	cmd, handled := CreateTab(m)

	if !handled {
		t.Error("Expected handled to be true")
	}
	if cmd == nil {
		t.Error("Expected reload command to be returned")
	}

	// Should have 2 tabs now
	if len(m.Tabs) != 2 {
		t.Errorf("Expected 2 tabs, got %d", len(m.Tabs))
	}

	// Active tab should be the new tab (index 1)
	if m.ActiveTab != 1 {
		t.Errorf("Expected active tab to be 1, got %d", m.ActiveTab)
	}

	// New tab should have current path
	if m.Tabs[1].Path != "/test2" {
		t.Errorf("Expected new tab path '/test2', got '%s'", m.Tabs[1].Path)
	}

	// First tab should have saved state
	if m.Tabs[0].Cursor != 5 {
		t.Errorf("Expected first tab cursor to be saved as 5, got %d", m.Tabs[0].Cursor)
	}
}

func TestCreateTab_AtLimit(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	// Create 9 tabs (the limit)
	m.Tabs = make([]state.Tab, 9)
	for i := 0; i < 9; i++ {
		m.Tabs[i] = state.Tab{
			Path:          "/test",
			SortMode:      sorting.SortDefault,
			SelectedPaths: make(map[string]bool),
		}
	}
	m.ActiveTab = 0

	_, handled := CreateTab(m)

	if !handled {
		t.Error("Expected handled to be true even at limit")
	}

	// Should still have 9 tabs
	if len(m.Tabs) != 9 {
		t.Errorf("Expected 9 tabs (no new tab created), got %d", len(m.Tabs))
	}
}

func TestSwitchTab_Success(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", Cursor: 1, Offset: 0, SortMode: sorting.SortName, SelectedPaths: make(map[string]bool)},
		{Path: "/test2", Cursor: 5, Offset: 3, SortMode: sorting.SortNewest, SelectedPaths: make(map[string]bool)},
		{Path: "/test3", Cursor: 0, Offset: 0, SortMode: sorting.SortSizeDesc, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 0
	m.Navigation.Path = "/test1"
	m.Navigation.Cursor = 10
	m.Navigation.Offset = 5
	m.Display.SortMode = sorting.SortName

	// Switch to tab 2 (index 1)
	cmd, handled := SwitchTab(m, 2)

	if !handled {
		t.Error("Expected handled to be true")
	}
	if cmd == nil {
		t.Error("Expected reload command to be returned")
	}

	// Active tab should be 1
	if m.ActiveTab != 1 {
		t.Errorf("Expected active tab to be 1, got %d", m.ActiveTab)
	}

	// State should be synced from tab 2
	if m.Navigation.Path != "/test2" {
		t.Errorf("Expected path '/test2', got '%s'", m.Navigation.Path)
	}
	if m.Navigation.Cursor != 5 {
		t.Errorf("Expected cursor 5, got %d", m.Navigation.Cursor)
	}
	if m.Navigation.Offset != 3 {
		t.Errorf("Expected offset 3, got %d", m.Navigation.Offset)
	}
	if m.Display.SortMode != sorting.SortNewest {
		t.Errorf("Expected sort mode SortNewest, got %v", m.Display.SortMode)
	}

	// Previous tab (tab 1) should have saved state
	if m.Tabs[0].Cursor != 10 {
		t.Errorf("Expected previous tab cursor saved as 10, got %d", m.Tabs[0].Cursor)
	}
	if m.Tabs[0].Offset != 5 {
		t.Errorf("Expected previous tab offset saved as 5, got %d", m.Tabs[0].Offset)
	}
}

func TestSwitchTab_Invalid(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{{Path: "/t1", SelectedPaths: make(map[string]bool)}}

	cmd, handled := SwitchTab(m, 5) // Non-existent
	if handled {
		t.Error("Expected handled false for invalid tab")
	}
	if cmd != nil {
		t.Error("Expected nil cmd")
	}
}

func TestCloseTab_Success(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
		{Path: "/test2", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
		{Path: "/test3", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 1 // Close middle tab

	cmd, handled := CloseTab(m)

	if !handled {
		t.Error("Expected handled to be true")
	}
	if cmd == nil {
		t.Error("Expected reload command to be returned")
	}

	// Should have 2 tabs now
	if len(m.Tabs) != 2 {
		t.Errorf("Expected 2 tabs after closing, got %d", len(m.Tabs))
	}

	// Active tab should still be 1 (which is now the old tab 3)
	if m.ActiveTab != 1 {
		t.Errorf("Expected active tab to be 1, got %d", m.ActiveTab)
	}

	// Remaining tabs should be tab 1 and tab 3
	if m.Tabs[0].Path != "/test1" {
		t.Error("First tab should be unchanged")
	}
	if m.Tabs[1].Path != "/test3" {
		t.Error("Second tab should be the old third tab")
	}
}

func TestCloseTab_Single(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{{Path: "/t1", SelectedPaths: make(map[string]bool)}}

	cmd, handled := CloseTab(m)
	if handled {
		t.Error("Expected handled false for single tab close")
	}
	if cmd != nil {
		t.Error("Expected nil cmd")
	}
}

func TestTabState(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{{Path: "/old", SelectedPaths: make(map[string]bool)}}
	m.ActiveTab = 0

	m.Navigation.Path = "/new"
	m.Navigation.Cursor = 10

	SaveTabState(m)
	if m.Tabs[0].Path != "/new" {
		t.Errorf("Expected tab path /new, got %s", m.Tabs[0].Path)
	}
	if m.Tabs[0].Cursor != 10 {
		t.Errorf("Expected tab cursor 10, got %d", m.Tabs[0].Cursor)
	}

	m.Navigation.Path = "/other"
	m.Navigation.Cursor = 0

	SyncTabToModel(m)
	if m.Navigation.Path != "/new" {
		t.Errorf("Expected path /new after sync, got %s", m.Navigation.Path)
	}
	if m.Navigation.Cursor != 10 {
		t.Errorf("Expected cursor 10 after sync, got %d", m.Navigation.Cursor)
	}
}

func TestSaveTabState(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{
		{Path: "/old", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 0
	m.Navigation.Path = "/new"
	m.Navigation.Cursor = 10
	m.Navigation.Offset = 5
	m.Display.SortMode = sorting.SortName
	m.Git.Branch = "main"
	m.Git.Root = "/repo"
	m.UI.InputActive = true
	m.Inputs.Mode = state.InputSearch
	m.Inputs.ActiveInput.SetValue("test")
	m.UI.SelectMode = true
	m.Navigation.SelectedPaths["/new/file1"] = true

	SaveTabState(m)

	// Verify all state was saved to the tab
	tab := m.Tabs[0]
	if tab.Path != "/new" {
		t.Errorf("Expected tab path '/new', got '%s'", tab.Path)
	}
	if tab.Cursor != 10 {
		t.Errorf("Expected tab cursor 10, got %d", tab.Cursor)
	}
	if tab.Offset != 5 {
		t.Errorf("Expected tab offset 5, got %d", tab.Offset)
	}
	if tab.SortMode != sorting.SortName {
		t.Errorf("Expected tab sort mode SortName, got %v", tab.SortMode)
	}
	if tab.GitBranch != "main" {
		t.Errorf("Expected tab git branch 'main', got '%s'", tab.GitBranch)
	}
	if !tab.Searching {
		t.Error("Expected tab searching to be true")
	}
	if tab.SearchQuery != "test" {
		t.Errorf("Expected tab search query 'test', got '%s'", tab.SearchQuery)
	}
	if !tab.SelectMode {
		t.Error("Expected tab select mode to be true")
	}
	if !tab.SelectedPaths["/new/file1"] {
		t.Error("Expected selected path to be saved")
	}
}

func TestSyncTabToModel(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{
		{
			Path:          "/test",
			Cursor:        10,
			Offset:        5,
			SortMode:      sorting.SortName,
			GitBranch:     "develop",
			GitRoot:       "/repo",
			Searching:     true,
			SearchQuery:   "query",
			SelectMode:    true,
			SelectedPaths: map[string]bool{"/test/file1": true},
		},
	}
	m.ActiveTab = 0
	m.Navigation.PathGen = 5

	SyncTabToModel(m)

	// Verify all state was synced from tab
	if m.Navigation.Path != "/test" {
		t.Errorf("Expected path '/test', got '%s'", m.Navigation.Path)
	}
	if m.Navigation.Cursor != 10 {
		t.Errorf("Expected cursor 10, got %d", m.Navigation.Cursor)
	}
	if m.Navigation.Offset != 5 {
		t.Errorf("Expected offset 5, got %d", m.Navigation.Offset)
	}
	if m.Display.SortMode != sorting.SortName {
		t.Errorf("Expected sort mode SortName, got %v", m.Display.SortMode)
	}
	if m.Git.Branch != "develop" {
		t.Errorf("Expected git branch 'develop', got '%s'", m.Git.Branch)
	}
	if !m.UI.InputActive || m.Inputs.Mode != state.InputSearch {
		t.Error("Expected input active and mode search to be true")
	}
	if m.Inputs.ActiveInput.Value() != "query" {
		t.Errorf("Expected search value 'query', got '%s'", m.Inputs.ActiveInput.Value())
	}
	if !m.UI.SelectMode {
		t.Error("Expected select mode to be true")
	}
	if !m.Navigation.SelectedPaths["/test/file1"] {
		t.Error("Expected selected path to be synced")
	}

	// PathGen should be incremented
	if m.Navigation.PathGen != 6 {
		t.Errorf("Expected PathGen to be incremented to 6, got %d", m.Navigation.PathGen)
	}
}

func TestSyncTabToModel_NotSearching(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Tabs = []state.Tab{
		{
			Path:      "/test",
			Searching: false,
		},
	}
	m.ActiveTab = 0
	m.UI.InputActive = true

	SyncTabToModel(m)

	if m.UI.InputActive {
		t.Error("Expected InputActive to be false after sync")
	}
}
