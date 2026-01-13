package update

import (
	"context"
	"os"
	"testing"

	"fm/internal/testutil"
	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleGoto(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.InputActive = true
	m.Inputs.Mode = state.InputGoto
	m.Inputs.ActiveInput.SetValue("/test/path")

	// Test local goto
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	m.FS.(*testutil.MockFileSystem).StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
		return &testutil.MockFileInfo{IsDirBool: true}, nil
	}
	_ = HandleGoto(msg, m)
	if m.UI.InputActive {
		t.Error("Expected goto input to close on enter")
	}
	if m.Navigation.Path != "/test/path" {
		t.Errorf("Expected path /test/path, got %s", m.Navigation.Path)
	}

	// Test remote goto
	m.UI.InputActive = true
	m.Inputs.Mode = state.InputGoto
	m.Inputs.ActiveInput.SetValue("user@host")
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	_ = HandleGoto(msg, m)
	if m.Remote.Host != "host" || m.Remote.User != "user" {
		t.Errorf("Expected host and user to be set, got %s@%s", m.Remote.User, m.Remote.Host)
	}
}
