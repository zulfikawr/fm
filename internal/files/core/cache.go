package core

import (
	"container/list"
	"sync"
	"time"
)

// CacheEntry holds a value and its metadata
type CacheEntry[V any] struct {
	Value   V
	Created time.Time
}

// SimpleCache implements a generic thread-safe size-limited cache with optional TTL.
type SimpleCache[K comparable, V any] struct {
	mu        sync.RWMutex
	capacity  int
	ttl       time.Duration
	cache     map[K]CacheEntry[V]
	order     *list.List
	protected map[K]bool
}

// NewSimpleCache creates a new generic cache.
func NewSimpleCache[K comparable, V any](capacity int, ttl time.Duration) *SimpleCache[K, V] {
	return &SimpleCache[K, V]{
		capacity:  capacity,
		ttl:       ttl,
		cache:     make(map[K]CacheEntry[V]),
		order:     list.New(),
		protected: make(map[K]bool),
	}
}

// Get retrieves a value from the cache if it hasn't expired.
func (sc *SimpleCache[K, V]) Get(key K) (V, bool) {
	sc.mu.RLock()
	entry, ok := sc.cache[key]
	sc.mu.RUnlock()

	if !ok {
		var zero V
		return zero, false
	}

	if sc.ttl > 0 && time.Since(entry.Created) > sc.ttl && !sc.isProtected(key) {
		sc.Delete(key)
		var zero V
		return zero, false
	}

	return entry.Value, true
}

// Put adds or updates a value in the cache.
func (sc *SimpleCache[K, V]) Put(key K, value V) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	entry, ok := sc.cache[key]
	if ok {
		sc.cache[key] = CacheEntry[V]{Value: value, Created: time.Now()}
		if entry.Created.IsZero() {
			// Dummy check
		}
		return
	}

	sc.cache[key] = CacheEntry[V]{Value: value, Created: time.Now()}
	sc.order.PushFront(key)

	if sc.capacity > 0 && sc.order.Len() > sc.capacity {
		// Find first non-protected item from the back to evict
		for e := sc.order.Back(); e != nil; e = e.Prev() {
			k := e.Value.(K)
			if !sc.protected[k] {
				sc.order.Remove(e)
				delete(sc.cache, k)
				break
			}
		}
	}
}

// Protect prevents a key from being evicted.
func (sc *SimpleCache[K, V]) Protect(key K) {
	if sc == nil {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.protected == nil {
		sc.protected = make(map[K]bool)
	}
	sc.protected[key] = true
}

// Unprotect allows a key to be evicted again.
func (sc *SimpleCache[K, V]) Unprotect(key K) {
	if sc == nil {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.protected == nil {
		return
	}
	delete(sc.protected, key)
}

func (sc *SimpleCache[K, V]) isProtected(key K) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	if sc.protected == nil {
		return false
	}
	return sc.protected[key]
}

// Delete removes a specific key from the cache.
func (sc *SimpleCache[K, V]) Delete(key K) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.cache, key)
	for e := sc.order.Front(); e != nil; e = e.Next() {
		if e.Value.(K) == key {
			sc.order.Remove(e)
			break
		}
	}
}

// Invalidate is an alias for Delete, to maintain compatibility with some callers.
func (sc *SimpleCache[K, V]) Invalidate(key K) {
	sc.Delete(key)
}

// Clear empties the cache.
func (sc *SimpleCache[K, V]) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.cache = make(map[K]CacheEntry[V])
	sc.order.Init()
}
