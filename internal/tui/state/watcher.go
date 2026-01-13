package state

import "github.com/fsnotify/fsnotify"

// WatcherState holds filesystem and git watching state
type WatcherState struct {
	Watcher     *fsnotify.Watcher // Filesystem watcher
	LastWatched string            // Last watched directory path
}
