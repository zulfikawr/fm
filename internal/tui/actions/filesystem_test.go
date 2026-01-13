package actions

import (
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestWatchDir(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/test/path"

	cmd := WatchDir(m)

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}
}

func TestWatchDir_NilWatcher(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Watcher.Watcher = nil

	cmd := WatchDir(m)

	if cmd != nil {
		t.Error("Expected no command when watcher is nil")
	}
}

func TestRestartWatcher(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Navigation.Path = "/new/path"

	// Create a watcher to be closed
	watcher, _ := fsnotify.NewWatcher()
	m.Watcher.Watcher = watcher

	cmd := RestartWatcher(m)

	if m.Watcher.Watcher == nil {
		t.Error("Expected new watcher to be created")
	}

	if cmd == nil {
		t.Error("Expected command to be returned")
	}
}

func TestHandleWatcherError(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	err := fsnotify.ErrClosed

	cmd := HandleWatcherError(m, err)
	if cmd == nil {
		t.Error("Expected command to be returned")
	}
}

func TestHandleWatcherClosed(t *testing.T) {
	m := tuitestutil.CreateTestModel()

	cmd := HandleWatcherClosed(m)
	if cmd == nil {
		t.Error("Expected command to be returned")
	}
}
