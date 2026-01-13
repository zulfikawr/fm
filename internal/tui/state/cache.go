package state

import "fm/internal/tui/cache"

// CacheState holds caching-related state
type CacheState struct {
	CursorMemory   *cache.SimpleCache           // Path -> Cursor index (LRU cache)
	OffsetMemory   *cache.SimpleCache           // Path -> Scroll offset (LRU cache)
	GitStatusCache map[string]map[string]string // Path -> Git status map (simple cache)
}
