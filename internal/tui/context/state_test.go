package context

import (
	"fm/internal/testutil"
	"testing"
)

func TestUIState(t *testing.T) {
	s := &UIState{}

	t.Run("StartInput", func(t *testing.T) {
		s.SettingsOpen = true
		s.StartInput()
		testutil.AssertEqual(t, true, s.InputActive, "Input should be active")
		testutil.AssertEqual(t, true, s.SettingsOpen, "Settings should NOT be closed automatically")
	})

	t.Run("Reset", func(t *testing.T) {
		s.InputActive = true
		s.Confirming = true
		s.Reset()
		testutil.AssertEqual(t, false, s.InputActive, "Input should be reset")
		testutil.AssertEqual(t, false, s.Confirming, "Confirming should be reset")
	})

	t.Run("Toggle Modals", func(t *testing.T) {
		s.Reset()
		s.ToggleSettings()
		testutil.AssertEqual(t, true, s.SettingsOpen, "Settings should be open")

		s.ToggleLogs()
		testutil.AssertEqual(t, true, s.LogOpen, "Logs should be open")
		testutil.AssertEqual(t, false, s.SettingsOpen, "Settings should be closed when logs open")
	})
}

func TestNavigationState(t *testing.T) {
	n := &NavigationState{
		SelectedPaths: make(map[string]bool),
	}

	t.Run("Selection", func(t *testing.T) {
		path := "/test/file"
		n.Select(path)
		testutil.AssertEqual(t, true, n.IsSelected(path), "Path should be selected")
		testutil.AssertEqual(t, 1, n.SelectedCount, "Count should be 1")

		n.ToggleSelection(path)
		testutil.AssertEqual(t, false, n.IsSelected(path), "Path should be deselected")
		testutil.AssertEqual(t, 0, n.SelectedCount, "Count should be 0")
	})
}

func TestClipboardState(t *testing.T) {
	c := &ClipboardState{}
	fs := testutil.NewMockFileSystem()

	t.Run("SetCopy", func(t *testing.T) {
		c.SetCopy(fs, []string{"/a", "/b"})
		testutil.AssertEqual(t, 2, len(c.Paths), "Should have 2 paths")
		testutil.AssertEqual(t, false, c.IsCut, "IsCut should be false")
		testutil.AssertEqual(t, "copy", c.Action, "Action should be copy")
	})

	t.Run("Clear", func(t *testing.T) {
		c.Clear()
		testutil.AssertEqual(t, 0, len(c.Paths), "Should have 0 paths")
		testutil.AssertEqual(t, "", c.Action, "Action should be empty")
	})
}

func TestProgressState(t *testing.T) {
	p := &ProgressState{}

	t.Run("Show and Update", func(t *testing.T) {
		p.Show("Working")
		testutil.AssertEqual(t, true, p.Visible, "Should be visible")
		testutil.AssertEqual(t, "Working", p.Label, "Label should match")

		p.Update(0.5)
		testutil.AssertEqual(t, 0.5, p.Percent, "Percent should match")

		p.Hide()
		testutil.AssertEqual(t, false, p.Visible, "Should be hidden")
	})
}

func TestLogState(t *testing.T) {
	s := &LogState{}

	t.Run("Add and Update", func(t *testing.T) {
		s.AddEntry(LogEntry{ID: "1", Message: "Msg 1"})
		testutil.AssertEqual(t, 1, len(s.Entries), "Should have 1 entry")

		s.UpdateStatus("1", StatusSuccess, LogSuccess, "Updated Msg", "Details")
		testutil.AssertEqual(t, StatusSuccess, s.Entries[0].Status, "Status should be success")
		testutil.AssertEqual(t, "Updated Msg", s.Entries[0].Message, "Message should be updated")
	})
}

func TestMessageState(t *testing.T) {
	s := &MessageState{}

	t.Run("Push and Pop", func(t *testing.T) {
		s.Push("Msg 1", false)
		testutil.AssertEqual(t, "Msg 1", s.Text, "Text should match")
		testutil.AssertEqual(t, 1, len(s.Stack), "Stack should have 1 item")

		s.Push("Msg 2", true)
		testutil.AssertEqual(t, "Msg 2", s.Text, "Text should match latest")

		s.Pop()
		testutil.AssertEqual(t, "Msg 2", s.Text, "Text should still be Msg 2 because Pop removes from FRONT (oldest)")

		s.Pop()
		testutil.AssertEqual(t, "", s.Text, "Text should be empty after popping all")
	})
}
