package context_test

import (
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/context"
)

func TestUIState(t *testing.T) {
	ui := &context.UIState{}

	t.Run("StartInput", func(t *testing.T) {
		ui.StartInput()
		testutil.AssertEqual(t, true, ui.InputActive, "InputActive should be true")
	})

	t.Run("StopInput", func(t *testing.T) {
		ui.StopInput()
		testutil.AssertEqual(t, false, ui.InputActive, "InputActive should be false")
	})

	t.Run("StartConfirming", func(t *testing.T) {
		ui.StartConfirming()
		testutil.AssertEqual(t, true, ui.Confirming, "Confirming should be true")
	})

	t.Run("StopConfirming", func(t *testing.T) {
		ui.StopConfirming()
		testutil.AssertEqual(t, false, ui.Confirming, "Confirming should be false")
	})

	t.Run("ToggleSettings", func(t *testing.T) {
		initial := ui.SettingsOpen
		ui.ToggleSettings()
		testutil.AssertEqual(t, !initial, ui.SettingsOpen, "SettingsOpen should be toggled")
	})

	t.Run("ToggleLogs", func(t *testing.T) {
		initial := ui.LogOpen
		ui.ToggleLogs()
		testutil.AssertEqual(t, !initial, ui.LogOpen, "LogOpen should be toggled")
	})

	t.Run("ToggleClipboard", func(t *testing.T) {
		initial := ui.ClipboardOpen
		ui.ToggleClipboard()
		testutil.AssertEqual(t, !initial, ui.ClipboardOpen, "ClipboardOpen should be toggled")
	})

	t.Run("Reset", func(t *testing.T) {
		ui.Confirming = true
		ui.Reset()
		testutil.AssertEqual(t, false, ui.Confirming, "Confirming should be reset to false")
	})
}

func TestNavigationState_Extended(t *testing.T) {
	nav := &context.NavigationState{}
	nav.FilteredItems = []core.Item{
		{Name: "..", State: core.ItemState{IsUp: true}, Path: "/"},
		{Name: "f1", Path: "/f1"},
	}

	t.Run("Selection Methods", func(t *testing.T) {
		nav.Select("/f1")
		testutil.AssertEqual(t, true, nav.IsSelected("/f1"), "Should be selected")
		testutil.AssertEqual(t, 1, nav.SelectedCount, "Count should be 1")
		testutil.AssertEqual(t, true, nav.FilteredItems[1].State.Selected, "FilteredItem should be visually selected")

		nav.ToggleSelection("/f1")
		testutil.AssertEqual(t, false, nav.IsSelected("/f1"), "Should be deselected after toggle")
		testutil.AssertEqual(t, false, nav.FilteredItems[1].State.Selected, "FilteredItem should be visually deselected")

		nav.SelectAll()
		testutil.AssertEqual(t, true, nav.IsSelected("/f1"), "Should be selected after SelectAll")
		testutil.AssertEqual(t, true, nav.FilteredItems[1].State.Selected, "FilteredItem should be visually selected after SelectAll")
		// ".." should not be selected
		testutil.AssertEqual(t, false, nav.IsSelected("/"), "IsUp item should not be selected")
	})
}

func TestOperationsState(t *testing.T) {
	ops := &context.OperationsState{}
	fs := testutil.NewMockFileSystem()

	t.Run("Clipboard", func(t *testing.T) {
		ops.Clipboard.SetCopy(fs, []string{"/a"})
		testutil.AssertEqual(t, 1, len(ops.Clipboard.Paths), "Should have 1 path")
		testutil.AssertEqual(t, false, ops.Clipboard.IsCut, "IsCut should be false")

		ops.Clipboard.SetCut(fs, []string{"/b"})
		testutil.AssertEqual(t, true, ops.Clipboard.IsCut, "IsCut should be true")

		ops.Clipboard.Clear()
		testutil.AssertEqual(t, 0, len(ops.Clipboard.Paths), "Should be empty after clear")
	})

	t.Run("Progress", func(t *testing.T) {
		ops.Progress.Show("label")
		testutil.AssertEqual(t, true, ops.Progress.Visible, "Should be visible")
		testutil.AssertEqual(t, "label", ops.Progress.Label, "Label should match")

		ops.Progress.Update(0.5)
		testutil.AssertEqual(t, 0.5, ops.Progress.Percent, "Percent should be 0.5")

		ops.Progress.Hide()
		testutil.AssertEqual(t, false, ops.Progress.Visible, "Should be hidden")
	})

	t.Run("Conflict", func(t *testing.T) {
		ops.Conflict.Set(context.ConflictParams{
			Source:       "/src",
			Destination:  "/dst",
			PendingItems: []string{"/src"},
			IsMove:       false,
			OpType:       "copy",
			LogID:        "log1",
		})
		testutil.AssertEqual(t, true, ops.Conflict.HasConflict(), "Should have conflict")
		testutil.AssertEqual(t, "/src", ops.Conflict.Source, "Source should match")

		ops.Conflict.Clear()
		testutil.AssertEqual(t, false, ops.Conflict.HasConflict(), "Should not have conflict after clear")
	})
}

func TestLogState(t *testing.T) {
	ls := &context.LogState{}

	t.Run("AddEntry", func(t *testing.T) {
		ls.AddEntry(context.LogEntry{ID: "1", Message: "m1"})
		testutil.AssertEqual(t, 1, len(ls.Entries), "Should have 1 entry")
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		ls.UpdateStatus("1", context.LogEntry{
			Status:  context.StatusSuccess,
			Level:   context.LogSuccess,
			Message: "updated",
			Details: "details",
		})
		testutil.AssertEqual(t, context.StatusSuccess, ls.Entries[0].Status, "Status should be updated")
		testutil.AssertEqual(t, "updated", ls.Entries[0].Message, "Message should be updated")
	})
}

func TestMessageState(t *testing.T) {
	ms := &context.MessageState{}

	t.Run("Push and Pop", func(t *testing.T) {
		ms.Push("msg1", false)
		testutil.AssertEqual(t, "msg1", ms.Text, "Text should match")
		testutil.AssertEqual(t, 1, len(ms.Stack), "Stack size should be 1")

		ms.Pop()
		testutil.AssertEqual(t, "", ms.Text, "Text should be empty after pop")
		testutil.AssertEqual(t, 0, len(ms.Stack), "Stack should be empty")
	})
}

func TestModel_Extended(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := context.NewModel(fs, "/test")

	t.Run("Input Handling", func(t *testing.T) {
		m.StartInput(context.InputSearch)
		testutil.AssertEqual(t, true, m.UI.InputActive, "InputActive should be true")
		testutil.AssertEqual(t, context.InputSearch, m.Inputs.Mode, "Mode should be InputSearch")

		m.StopInput(true)
		testutil.AssertEqual(t, false, m.UI.InputActive, "InputActive should be false")
		testutil.AssertEqual(t, context.InputNone, m.Inputs.Mode, "Mode should be InputNone")
	})

	t.Run("Viewport", func(t *testing.T) {
		m.Display.Height = 20
		m.SyncViewportHeight()
		// App Header(1) + App Footer(1) = 2. 20-2 = 18.
		if m.Display.ViewportHeight != 18 {
			t.Errorf("expected viewport height 18, got %d", m.Display.ViewportHeight)
		}
	})
}
