package core

import (
	"testing"
	"time"
)

func TestSimpleCache(t *testing.T) {
	cache := NewSimpleCache[string, string](2, time.Second)

	t.Run("Put and Get", func(t *testing.T) {
		cache.Put("key1", "val1")
		val, ok := cache.Get("key1")
		if !ok || val != "val1" {
			t.Errorf("Get failed, expected val1, got %v", val)
		}
	})

	t.Run("Eviction", func(t *testing.T) {
		cache.Put("key1", "val1")
		cache.Put("key2", "val2")
		cache.Put("key3", "val3") // key1 should be evicted

		_, ok := cache.Get("key1")
		if ok {
			t.Error("key1 should have been evicted")
		}
	})

	t.Run("TTL", func(t *testing.T) {
		cache := NewSimpleCache[string, string](2, 10*time.Millisecond)
		cache.Put("key1", "val1")
		time.Sleep(20 * time.Millisecond)
		_, ok := cache.Get("key1")
		if ok {
			t.Error("key1 should have expired")
		}
	})

	t.Run("Invalidate", func(t *testing.T) {
		cache.Put("key1", "val1")
		cache.Invalidate("key1")
		_, ok := cache.Get("key1")
		if ok {
			t.Error("key1 should have been invalidated")
		}
	})
}
