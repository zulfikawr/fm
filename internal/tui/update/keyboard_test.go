package update

import (
	"testing"

	"fm/internal/files/core"
	"fm/internal/sshutil"
	"fm/internal/testutil"
	"fm/internal/tui/state"
	tuitestutil "fm/internal/tui/testutil"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

func TestHandleKeyMsg_Quit(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.FS.(*testutil.MockFileSystem).IsLocalFunc = func() bool { return true }

	// Create a real watcher for the test so Close doesn't panic
	watcher, _ := fsnotify.NewWatcher()
	m.Watcher.Watcher = watcher

	// Test 'q'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	cmd := HandleKeyMsg(m, msg)
	if cmd == nil {
		t.Fatal("Expected tea.Quit command, got nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("Expected tea.QuitMsg, got %T", cmd())
	}

	// Test 'ctrl+c'
	msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	cmd = HandleKeyMsg(m, msg)
	if cmd == nil {
		t.Fatal("Expected tea.Quit command, got nil")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("Expected tea.QuitMsg, got %T", cmd())
	}
}

func TestHandleKeyMsg_Escape(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.SelectMode = true
	m.Navigation.SelectedCount = 1
	m.Navigation.Items = []core.Item{{Name: "f1", Path: "/f1", Selected: true}}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_ = HandleKeyMsg(m, msg)

	if m.UI.SelectMode {
		t.Error("Expected SelectMode to be false")
	}
	if m.Navigation.Items[0].Selected {
		t.Error("Expected item to be deselected")
	}
}

func TestHandleKeyMsg_InputModes(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.InputActive = true
	m.Inputs.ActiveInput = textinput.New()
	m.Inputs.ActiveInput.Focus()

	modes := []struct {
		mode state.InputMode
		key  string
	}{
		{state.InputSearch, "/"},
		{state.InputRename, "n"},
		{state.InputGoto, "g"},
		{state.InputAuth, "p"},
	}

	for _, tt := range modes {
		m.Inputs.Mode = tt.mode
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
		_ = HandleKeyMsg(m, msg)
		if m.Inputs.ActiveInput.Value() != "a" {
			t.Errorf("Mode %v: Expected input value 'a', got '%s'", tt.mode, m.Inputs.ActiveInput.Value())
		}
		m.Inputs.ActiveInput.SetValue("")
	}
}

func TestHandleKeyMsg_HostConfirm(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.UI.HostConfirm = true
	resolve := make(chan bool, 1)
	m.Remote.HostConfirmReq = &sshutil.HostConfirmRequest{
		Resolve: resolve,
	}

	// Test 'y'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	_ = HandleKeyMsg(m, msg)
	if !<-resolve {
		t.Error("Expected resolve to be true")
	}
	if m.UI.HostConfirm {
		t.Error("Expected HostConfirm to be false")
	}

	// Test 'n'
	m.UI.HostConfirm = true
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	_ = HandleKeyMsg(m, msg)
	if <-resolve {
		t.Error("Expected resolve to be false")
	}

	// Test 'Y'
	m.UI.HostConfirm = true
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")}
	_ = HandleKeyMsg(m, msg)
	if !<-resolve {
		t.Error("Expected resolve true for 'Y'")
	}

	// Test 'N'
	m.UI.HostConfirm = true
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")}
	_ = HandleKeyMsg(m, msg)
	if <-resolve {
		t.Error("Expected resolve false for 'N'")
	}
}
