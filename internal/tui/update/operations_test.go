package update

import (
	"testing"

	"fm/internal/files"
	"fm/internal/tui/commands"
)

func TestHandleProgress(t *testing.T) {
	m := createTestModel()
	m.Operations.Progress.Visible = false

	progChan := make(chan files.Progress, 1)
	progChan <- files.Progress{
		Percent: 0.5,
		Label:   "Copying...",
	}

	msg := commands.ProgressMsg{
		Percent: 0.5,
		Label:   "Copying file.txt...",
		Channel: progChan,
	}

	cmd := HandleProgress(m, msg)

	if !m.Operations.Progress.Visible {
		t.Error("Expected progress to be visible")
	}
	if m.Operations.Progress.Percent != 0.5 {
		t.Errorf("Expected progress percent 0.5, got %f", m.Operations.Progress.Percent)
	}
	if m.Operations.Progress.Label != "Copying file.txt..." {
		t.Errorf("Expected progress label 'Copying file.txt...', got '%s'", m.Operations.Progress.Label)
	}
	if cmd == nil {
		t.Error("Expected command to continue listening to progress")
	}

	close(progChan)
}

func TestHandleOperationFinished_SingleItem(t *testing.T) {
	m := createTestModel()
	m.Operations.Progress.Visible = true
	m.Operations.ProcessingItems["/test/file1.txt"] = true
	m.Operations.SelectedPaths["/test/file1.txt"] = true
	m.UI.SelectMode = true

	msg := commands.OperationFinishedMsg{
		Paths: []string{"/test/file1.txt"},
	}

	HandleOperationFinished(m, msg)

	if m.Operations.Progress.Visible {
		t.Error("Expected progress to be hidden after operation finished")
	}
	if m.Operations.ProcessingItems["/test/file1.txt"] {
		t.Error("Expected item to be removed from processing items")
	}
	if m.Operations.SelectedPaths["/test/file1.txt"] {
		t.Error("Expected item to be removed from selected paths")
	}
	if m.UI.SelectMode {
		t.Error("Expected select mode to be disabled when no items are selected")
	}
}

func TestHandleOperationFinished_MultipleItems(t *testing.T) {
	m := createTestModel()
	m.Operations.Progress.Visible = true
	m.Operations.ProcessingItems["/test/file1.txt"] = true
	m.Operations.ProcessingItems["/test/file2.txt"] = true
	m.Operations.ProcessingItems["/test/file3.txt"] = true
	m.Operations.SelectedPaths["/test/file1.txt"] = true
	m.Operations.SelectedPaths["/test/file2.txt"] = true
	m.Operations.SelectedPaths["/test/file3.txt"] = true
	m.Operations.SelectedPaths["/test/file4.txt"] = true // This one not in finished list
	m.UI.SelectMode = true

	msg := commands.OperationFinishedMsg{
		Paths: []string{
			"/test/file1.txt",
			"/test/file2.txt",
			"/test/file3.txt",
		},
	}

	HandleOperationFinished(m, msg)

	if m.Operations.Progress.Visible {
		t.Error("Expected progress to be hidden")
	}
	if len(m.Operations.ProcessingItems) != 0 {
		t.Errorf("Expected all finished items removed from processing, got %d items", len(m.Operations.ProcessingItems))
	}
	if len(m.Operations.SelectedPaths) != 1 {
		t.Errorf("Expected 1 selected path remaining, got %d", len(m.Operations.SelectedPaths))
	}
	if !m.Operations.SelectedPaths["/test/file4.txt"] {
		t.Error("Expected file4.txt to still be selected")
	}
	if !m.UI.SelectMode {
		t.Error("Expected select mode to remain enabled since there's still a selected item")
	}
}

func TestHandleOperationFinished_AllItemsCleared(t *testing.T) {
	m := createTestModel()
	m.Operations.ProcessingItems["/test/file1.txt"] = true
	m.Operations.SelectedPaths["/test/file1.txt"] = true
	m.UI.SelectMode = true

	msg := commands.OperationFinishedMsg{
		Paths: []string{"/test/file1.txt"},
	}

	HandleOperationFinished(m, msg)

	if m.UI.SelectMode {
		t.Error("Expected select mode to be disabled when all selected items are finished")
	}
	if len(m.Operations.SelectedPaths) != 0 {
		t.Error("Expected all selected paths to be cleared")
	}
}
