package state

import (
	"fm/internal/files/sorting"
	"fm/internal/testutil"
	"testing"
)

func TestTab(t *testing.T) {
	mockFS := testutil.NewMockFileSystem()

	t.Run("NewTab", func(t *testing.T) {
		tab := NewTab(mockFS, "/test", sorting.SortName)
		if tab.Path != "/test" || tab.SortMode != sorting.SortName || tab.SelectedPaths == nil || tab.FS == nil {
			t.Errorf("NewTab failed: %+v", tab)
		}
	})

	t.Run("Selection", func(t *testing.T) {
		tab := NewTab(mockFS, "/", sorting.SortDefault)
		tab.SelectedPaths["/p1"] = true

		if tab.SelectedCount() != 1 {
			t.Errorf("SelectedCount failed: %d", tab.SelectedCount())
		}
		if !tab.IsSelected("/p1") {
			t.Error("IsSelected failed for true")
		}
		if tab.IsSelected("/p2") {
			t.Error("IsSelected failed for false")
		}
	})
}

func TestModelTabs(t *testing.T) {
	mockFS := testutil.NewMockFileSystem()
	m := &Model{FS: mockFS}

	// Test AddTab
	m.AddTab("/test")
	if len(m.Tabs) != 1 || m.ActiveTab != 0 {
		t.Error("Expected 1 tab at index 0")
	}
	if m.Tabs[0].FS == nil {
		t.Error("Expected FS to be set in tab")
	}

	// Test AddTab with remote info
	m.Remote.User = "user"
	m.Remote.Host = "host"
	m.AddTab("/remote")
	if len(m.Tabs) != 2 {
		t.Error("Expected 2 tabs")
	}
	if m.Tabs[1].RemoteUser != "user" || m.Tabs[1].RemoteHost != "host" {
		t.Errorf("Remote info not preserved: got %s@%s", m.Tabs[1].RemoteUser, m.Tabs[1].RemoteHost)
	}

	// Test SwitchTab
	m.AddTab("/test2")
	if !m.SwitchTab(1) {
		t.Error("Expected switch to tab 1")
	}
	if m.ActiveTab != 0 {
		t.Errorf("Expected active tab 0, got %d", m.ActiveTab)
	}

	// Test CloseActiveTab
	if !m.CloseActiveTab() {
		t.Error("Expected to close active tab")
	}
	if len(m.Tabs) != 2 {
		t.Errorf("Expected 2 tabs remaining, got %d", len(m.Tabs))
	}
}
