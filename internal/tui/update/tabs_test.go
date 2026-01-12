package update

import (
	"testing"

	"fm/internal/files/sorting"
	"fm/internal/tui/actions"
	"fm/internal/tui/state"
)

func TestHandleCreateTab_Success(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 0
	m.Navigation.Path = "/test2"
	m.Navigation.Cursor = 5
	m.Navigation.Offset = 3

	cmd, handled := HandleCreateTab(m)

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

func TestHandleCreateTab_AtLimit(t *testing.T) {
	m := createTestModel()

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

	_, handled := HandleCreateTab(m)

	if !handled {
		t.Error("Expected handled to be true even at limit")
	}

	// Should still have 9 tabs
	if len(m.Tabs) != 9 {
		t.Errorf("Expected 9 tabs (no new tab created), got %d", len(m.Tabs))
	}
}

func TestHandleCreateTab_NoExistingTabs(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{} // No tabs
	m.Navigation.Path = "/test"

	_, handled := HandleCreateTab(m)

	if !handled {
		t.Error("Expected handled to be true")
	}

	// Should create initial tab and then add new one
	if len(m.Tabs) < 1 {
		t.Fatal("Expected at least one tab to be created")
	}

	// First tab should be created from current state
	if m.Tabs[0].Path != "/test" {
		t.Errorf("Expected first tab path '/test', got '%s'", m.Tabs[0].Path)
	}
}

func TestHandleSwitchTab_Success(t *testing.T) {
	m := createTestModel()
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
	cmd, handled := HandleSwitchTab(m, 2)

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

func TestHandleSwitchTab_SingleTab(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 0

	// Try to switch tabs when there's only one
	cmd, handled := HandleSwitchTab(m, 2)

	if handled {
		t.Error("Expected handled to be false with only one tab")
	}
	if cmd != nil {
		t.Error("Expected no command with only one tab")
	}

	// Active tab should still be 0
	if m.ActiveTab != 0 {
		t.Error("Active tab should not change")
	}
}

func TestHandleSwitchTab_InvalidTabNumber(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
		{Path: "/test2", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 0

	// Try to switch to tab 5 (doesn't exist)
	cmd, handled := HandleSwitchTab(m, 5)

	if handled {
		t.Error("Expected handled to be false for invalid tab number")
	}
	if cmd != nil {
		t.Error("Expected no command for invalid tab number")
	}

	// Active tab should still be 0
	if m.ActiveTab != 0 {
		t.Error("Active tab should not change")
	}
}

func TestHandleSwitchTab_ZeroTabNumber(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
		{Path: "/test2", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 0

	// Try to switch to tab 0 (invalid)
	cmd, handled := HandleSwitchTab(m, 0)

	if handled {
		t.Error("Expected handled to be false for tab number 0")
	}
	if cmd != nil {
		t.Error("Expected no command for tab number 0")
	}
}

func TestHandleCloseTab_Success(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
		{Path: "/test2", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
		{Path: "/test3", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 1 // Close middle tab

	cmd, handled := HandleCloseTab(m)

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

func TestHandleCloseTab_LastTab(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
		{Path: "/test2", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 1 // Close last tab

	_, handled := HandleCloseTab(m)

	if !handled {
		t.Error("Expected handled to be true")
	}

	// Should have 1 tab now
	if len(m.Tabs) != 1 {
		t.Errorf("Expected 1 tab after closing, got %d", len(m.Tabs))
	}

	// Active tab should be adjusted to 0
	if m.ActiveTab != 0 {
		t.Errorf("Expected active tab to be adjusted to 0, got %d", m.ActiveTab)
	}
}

func TestHandleCloseTab_OnlyOneTab(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 0

	// Try to close the only tab
	_, handled := HandleCloseTab(m)

	if handled {
		t.Error("Expected handled to be false when trying to close the only tab")
	}

	// Should still have 1 tab
	if len(m.Tabs) != 1 {
		t.Error("Tab should not be closed when it's the only one")
	}
}

func TestHandleCloseTab_InvalidActiveTab(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test1", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
		{Path: "/test2", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 5 // Invalid index

	_, handled := HandleCloseTab(m)

	if handled {
		t.Error("Expected handled to be false for invalid active tab")
	}

	// Active tab should be adjusted to 0
	if m.ActiveTab != 0 {
		t.Errorf("Expected active tab to be adjusted to 0, got %d", m.ActiveTab)
	}

	// Tabs should be unchanged
	if len(m.Tabs) != 2 {
		t.Error("Tabs should not change with invalid active tab")
	}
}

func TestSaveTabState(t *testing.T) {
	m := createTestModel()
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
	m.Operations.SelectedPaths["/new/file1"] = true

	actions.SaveTabState(m)

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

func TestSaveTabState_InvalidActiveTab(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 5 // Invalid
	m.Navigation.Path = "/new"

	// Should not panic
	actions.SaveTabState(m)

	// Tab should be unchanged
	if m.Tabs[0].Path != "/test" {
		t.Error("Tab should not be modified with invalid active tab")
	}
}

func TestSyncTabToModel(t *testing.T) {
	m := createTestModel()
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

	actions.SyncTabToModel(m)

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
	if !m.Operations.SelectedPaths["/test/file1"] {
		t.Error("Expected selected path to be synced")
	}

	// PathGen should be incremented
	if m.Navigation.PathGen != 6 {
		t.Errorf("Expected PathGen to be incremented to 6, got %d", m.Navigation.PathGen)
	}
}

func TestSyncTabToModel_InvalidActiveTab(t *testing.T) {
	m := createTestModel()
	m.Tabs = []state.Tab{
		{Path: "/test", SortMode: sorting.SortDefault, SelectedPaths: make(map[string]bool)},
	}
	m.ActiveTab = 5 // Invalid
	m.Navigation.Path = "/old"
	m.Navigation.PathGen = 1

	// Should not panic
	actions.SyncTabToModel(m)

	// State should be unchanged
	if m.Navigation.Path != "/old" {
		t.Error("Path should not change with invalid active tab")
	}
	if m.Navigation.PathGen != 1 {
		t.Error("PathGen should not change with invalid active tab")
	}
}

func TestParseTabNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"alt+1", 1},
		{"alt+2", 2},
		{"alt+5", 5},
		{"alt+9", 9},
	}

	for _, tt := range tests {
		result := ParseTabNumber(tt.input)
		if result != tt.expected {
			t.Errorf("ParseTabNumber(%s) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
