package context_test

import (
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/context"
)

func TestModel_TabManagement(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := context.NewModel(fs, "/home")

	t.Run("Initial State", func(t *testing.T) {
		testutil.AssertEqual(t, 1, len(m.Tabs), "Should have 1 tab initially")
		testutil.AssertEqual(t, 0, m.ActiveTab, "Active tab should be 0")
	})

	t.Run("AddTab", func(t *testing.T) {
		m.AddTab("/tmp")
		testutil.AssertEqual(t, 2, len(m.Tabs), "Should have 2 tabs")
		testutil.AssertEqual(t, "/tmp", m.Tabs[1].Path, "Second tab path should be /tmp")
	})

	t.Run("SwitchTab", func(t *testing.T) {
		success := m.SwitchTab(2)
		testutil.AssertEqual(t, true, success, "SwitchTab should succeed")
		testutil.AssertEqual(t, 1, m.ActiveTab, "Active tab should be 1")

		success = m.SwitchTab(5) // Invalid
		testutil.AssertEqual(t, false, success, "SwitchTab to invalid index should fail")
	})

	t.Run("CloseActiveTab", func(t *testing.T) {
		m := context.NewModel(fs, "/test")
		m.AddTab("/tmp")
		m.ActiveTab = 1

		success := m.CloseActiveTab()
		testutil.AssertEqual(t, true, success, "CloseActiveTab should succeed")
		testutil.AssertEqual(t, 1, len(m.Tabs), "Should have 1 tab left")
		testutil.AssertEqual(t, 0, m.ActiveTab, "Active tab should be 0")

		success = m.CloseActiveTab()
		testutil.AssertEqual(t, false, success, "Should not be able to close the last tab")
	})

	t.Run("AddTab limit", func(t *testing.T) {
		m := context.NewModel(fs, "/test")
		for i := 0; i < 15; i++ {
			m.AddTab("/path")
		}
		testutil.AssertEqual(t, 9, len(m.Tabs), "Should not exceed 9 tabs")
	})
}

func TestModel_ClearSelection(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := context.NewModel(fs, "/home")

	m.Navigation.Select("/home/file1")
	m.Navigation.SelectMode = true

	testutil.AssertEqual(t, 1, m.Navigation.SelectedCount(), "Should have 1 selected item")

	m.ClearSelection()

	testutil.AssertEqual(t, 0, m.Navigation.SelectedCount(), "Should have 0 selected items after clear")
	testutil.AssertEqual(t, false, m.Navigation.SelectMode, "SelectMode should be false")
}
