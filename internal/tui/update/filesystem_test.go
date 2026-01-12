package update

import (
	"strings"
	"testing"

	"fm/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

func TestHandleWindowSize(t *testing.T) {
	m := createTestModel()
	m.Display.Width = 0
	m.Display.Height = 0

	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	}

	cmd := HandleWindowSize(m, msg)

	if m.Display.Width != 100 {
		t.Errorf("Expected width 100, got %d", m.Display.Width)
	}
	if m.Display.Height != 50 {
		t.Errorf("Expected height 50, got %d", m.Display.Height)
	}
	if cmd != nil {
		t.Error("Expected no command from window size handler")
	}
}

func TestWatchDir(t *testing.T) {
	m := createTestModel()
	m.Navigation.Path = "/test/path"

	cmd := WatchDir(m)

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}

	// WatchDir returns a command that will watch when executed
	// LastWatched is set by the commands.WatchDir implementation, not by update.WatchDir
}

func TestWatchDir_NilWatcher(t *testing.T) {
	m := createTestModel()
	m.Watcher.Watcher = nil

	cmd := WatchDir(m)

	if cmd != nil {
		t.Error("Expected no command when watcher is nil")
	}
}

func TestHandleWatchEvent_Create(t *testing.T) {
	m := createTestModel()
	m.Navigation.Path = "/test"

	msg := commands.WatchEventMsg{
		Event: fsnotify.Event{
			Name: "/test/newfile.txt",
			Op:   fsnotify.Create,
		},
	}

	cmd := HandleWatchEvent(m, msg)

	if cmd == nil {
		t.Fatal("Expected reload command after create event")
	}

	// HandleWatchEvent returns a batch command (Reload + WatchDir)
	// We can't easily test the batch contents without executing multiple layers
	// Just verify a command was returned
}

func TestHandleWatchEvent_Write(t *testing.T) {
	m := createTestModel()
	m.Navigation.Path = "/test"

	msg := commands.WatchEventMsg{
		Event: fsnotify.Event{
			Name: "/test/file.txt",
			Op:   fsnotify.Write,
		},
	}

	cmd := HandleWatchEvent(m, msg)

	if cmd == nil {
		t.Fatal("Expected reload command after write event")
	}
}

func TestHandleWatchEvent_Remove(t *testing.T) {
	m := createTestModel()
	m.Navigation.Path = "/test"

	msg := commands.WatchEventMsg{
		Event: fsnotify.Event{
			Name: "/test/file.txt",
			Op:   fsnotify.Remove,
		},
	}

	cmd := HandleWatchEvent(m, msg)

	if cmd == nil {
		t.Fatal("Expected reload command after remove event")
	}
}

func TestHandleWatchEvent_Rename(t *testing.T) {
	m := createTestModel()
	m.Navigation.Path = "/test"

	msg := commands.WatchEventMsg{
		Event: fsnotify.Event{
			Name: "/test/oldfile.txt",
			Op:   fsnotify.Rename,
		},
	}

	cmd := HandleWatchEvent(m, msg)

	if cmd == nil {
		t.Fatal("Expected reload command after rename event")
	}
}

func TestHandleWatchEvent_Chmod(t *testing.T) {
	m := createTestModel()
	m.Navigation.Path = "/test"

	msg := commands.WatchEventMsg{
		Event: fsnotify.Event{
			Name: "/test/file.txt",
			Op:   fsnotify.Chmod,
		},
	}

	cmd := HandleWatchEvent(m, msg)

	// Chmod events should trigger reload
	if cmd == nil {
		t.Fatal("Expected reload command after chmod event")
	}
}

func TestHandleWatcherError(t *testing.T) {
	m := createTestModel()

	testErr := fsnotify.ErrClosed
	msg := commands.WatcherErrorMsg{
		Err: testErr,
	}

	errorLogged := false
	logError := func(err error, ctx string) tea.Cmd {
		errorLogged = true
		// Check if error message contains the original error
		if err != nil && !strings.Contains(err.Error(), testErr.Error()) {
			t.Errorf("Expected error containing %v, got %v", testErr, err)
		}
		return nil
	}

	setMsgCalled := false
	setMsg := func(s string) tea.Cmd {
		setMsgCalled = true
		if s != "Watcher error: restarting..." {
			t.Errorf("Unexpected message: %s", s)
		}
		return nil
	}

	cmd := HandleWatcherError(m, msg, logError, setMsg)

	if !errorLogged {
		t.Error("Expected error to be logged")
	}
	if !setMsgCalled {
		t.Error("Expected message to be set")
	}
	if cmd == nil {
		t.Error("Expected command to be returned")
	}
}

func TestHandleWatcherClosed(t *testing.T) {
	m := createTestModel()

	_ = commands.WatcherClosedMsg{}

	setMsgCalled := false
	setMsg := func(s string) tea.Cmd {
		setMsgCalled = true
		if s != "Watcher closed: restarting..." {
			t.Errorf("Unexpected message: %s", s)
		}
		return nil
	}

	cmd := HandleWatcherClosed(m, setMsg)

	if !setMsgCalled {
		t.Error("Expected message to be set")
	}
	if cmd == nil {
		t.Error("Expected command to be returned")
	}
}

func TestRestartWatcher(t *testing.T) {
	m := createTestModel()
	m.Navigation.Path = "/new/path"

	// Create a watcher to be closed
	watcher, _ := fsnotify.NewWatcher()
	m.Watcher.Watcher = watcher

	cmd := RestartWatcher(m)

	// The old watcher should be closed (we can't easily verify this without checking closed state)
	// A new watcher should be created
	if m.Watcher.Watcher == nil {
		t.Error("Expected new watcher to be created")
	}

	// Should return a command to watch the directory
	if cmd == nil {
		t.Error("Expected command to be returned")
	}
}

func TestRestartWatcher_NilWatcher(t *testing.T) {
	m := createTestModel()
	m.Watcher.Watcher = nil

	cmd := RestartWatcher(m)

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}

	// Execute the command - RestartWatcher modifies m as a side effect
	_ = cmd()

	// Should create a new watcher even if old one was nil
	if m.Watcher.Watcher == nil {
		t.Error("Expected new watcher to be created")
	}
}
