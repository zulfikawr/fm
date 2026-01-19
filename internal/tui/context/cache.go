package context

import (
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/zulfikawr/fm/internal/files/core"
)

// --- Cache State ---

// CacheState holds caching-related state
type CacheState struct {
	CursorMemory   *core.SimpleCache[string, int]               // Path -> Cursor index
	OffsetMemory   *core.SimpleCache[string, int]               // Path -> Scroll offset
	GitStatusCache *core.SimpleCache[string, map[string]string] // Path -> Git status map
	GitRootCache   *core.SimpleCache[string, string]            // Path -> Git root path
	ItemCache      *core.SimpleCache[string, []core.Item]       // Path -> Formatted items
}

// --- Watcher State ---

// WatcherState holds filesystem and git watching state
type WatcherState struct {
	Watcher       *fsnotify.Watcher // Filesystem watcher
	LastWatched   string            // Last watched directory path
	IsListening   bool              // Whether we are currently listening for events
	IsRemote      bool              // Whether the current listener is for a remote filesystem
	DebounceTimer *time.Timer       // Timer for debouncing events
}
