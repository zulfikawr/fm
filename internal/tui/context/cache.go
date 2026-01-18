package context

import (
	"container/list"
	"time"

	"github.com/fsnotify/fsnotify"

	"fm/internal/files/core"
)

// CacheEntry holds a value and its metadata
type CacheEntry[V any] struct {
	Value   V
	Created time.Time
}

// SimpleCache implements a generic size-limited cache with optional TTL
type SimpleCache[K comparable, V any] struct {
	capacity  int
	ttl       time.Duration
	cache     map[K]CacheEntry[V]
	order     *list.List
	protected map[K]bool
}

// NewSimpleCache creates a new generic cache
func NewSimpleCache[K comparable, V any](capacity int, ttl time.Duration) *SimpleCache[K, V] {
	return &SimpleCache[K, V]{
		capacity:  capacity,
		ttl:       ttl,
		cache:     make(map[K]CacheEntry[V]),
		order:     list.New(),
		protected: make(map[K]bool),
	}
}

// Get retrieves a value from the cache if it hasn't expired
func (c *SimpleCache[K, V]) Get(key K) (V, bool) {
	entry, ok := c.cache[key]
	if !ok {
		var zero V
		return zero, false
	}

	if c.ttl > 0 && time.Since(entry.Created) > c.ttl && !c.protected[key] {
		c.Delete(key)
		var zero V
		return zero, false
	}

	return entry.Value, true
}

// Put adds or updates a value in the cache
func (c *SimpleCache[K, V]) Put(key K, value V) {
	if _, ok := c.cache[key]; ok {
		c.cache[key] = CacheEntry[V]{Value: value, Created: time.Now()}
		return
	}

	c.cache[key] = CacheEntry[V]{Value: value, Created: time.Now()}
	c.order.PushFront(key)

	if c.order.Len() > c.capacity {
		// Find first non-protected item from the back to evict
		for e := c.order.Back(); e != nil; e = e.Prev() {
			k := e.Value.(K)
			if !c.protected[k] {
				c.order.Remove(e)
				delete(c.cache, k)
				break
			}
		}
	}
}

// Protect prevents a key from being evicted
func (c *SimpleCache[K, V]) Protect(key K) {
	if c == nil {
		return
	}
	if c.protected == nil {
		c.protected = make(map[K]bool)
	}
	c.protected[key] = true
}

// Unprotect allows a key to be evicted again
func (c *SimpleCache[K, V]) Unprotect(key K) {
	if c == nil || c.protected == nil {
		return
	}
	delete(c.protected, key)
}

// Delete removes a specific key from the cache
func (c *SimpleCache[K, V]) Delete(key K) {
	delete(c.cache, key)
	for e := c.order.Front(); e != nil; e = e.Next() {
		if e.Value.(K) == key {
			c.order.Remove(e)
			break
		}
	}
}

// Clear empties the cache
func (c *SimpleCache[K, V]) Clear() {
	c.cache = make(map[K]CacheEntry[V])
	c.order.Init()
}

// --- Cache State ---

// CacheState holds caching-related state
type CacheState struct {
	CursorMemory   *SimpleCache[string, int]               // Path -> Cursor index
	OffsetMemory   *SimpleCache[string, int]               // Path -> Scroll offset
	GitStatusCache *SimpleCache[string, map[string]string] // Path -> Git status map
	GitRootCache   *SimpleCache[string, string]            // Path -> Git root path
	ItemCache      *SimpleCache[string, []core.Item]       // Path -> Formatted items
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
