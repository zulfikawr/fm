package update

import (
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/commands"

	"github.com/fsnotify/fsnotify"
)

func TestHandleFileSystemMsg(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test"
	m.Navigation.PathGen = 1
	m.UI.Loading = true

	// Test LoadedItemsMsg
	msg := commands.LoadedItemsMsg{
		Generation: 1,
		Path:       "/test",
		Items:      []core.Item{{Name: "f1"}},
	}
	_ = HandleFileSystemMsg(m, msg)
	if m.UI.Loading {
		t.Error("Expected loading to be false after LoadedItemsMsg")
	}
	if len(m.Navigation.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(m.Navigation.Items))
	}
}

func TestHandleWatchEvent(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test"

	msg := commands.WatchEventMsg{
		Event: fsnotify.Event{Name: "/test/file.txt", Op: fsnotify.Write},
	}

	cmd := HandleWatchEvent(m, msg)
	if cmd == nil {
		t.Error("Expected reload command for watch event")
	}

	// Git watch event
	msgGit := commands.WatchEventMsg{
		Event: fsnotify.Event{Name: "/test/.git/index", Op: fsnotify.Write},
	}
	cmd = HandleWatchEvent(m, msgGit)
	if cmd == nil {
		t.Error("Expected git status fetch command")
	}
}
