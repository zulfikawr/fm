package core

import (
	"os"
	"testing"
	"time"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestSimpleCache(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := NewSimpleCache[string, []os.FileInfo](10, ttl)

	entries := []os.FileInfo{&testutil.MockFileInfo{FName: "file1"}}
	path := "/test/path"

	t.Run("Put and Get", func(t *testing.T) {
		cache.Put(path, entries)
		got, ok := cache.Get(path)
		if !ok {
			t.Fatal("Expected cache hit")
		}
		testutil.AssertEqual(t, 1, len(got), "Should have 1 entry")
	})

	t.Run("Expiration", func(t *testing.T) {
		cache.Put(path, entries)
		time.Sleep(ttl + 10*time.Millisecond)
		_, ok := cache.Get(path)
		if ok {
			t.Error("Expected cache miss after expiration")
		}
	})

	t.Run("Invalidate", func(t *testing.T) {
		cache.Put(path, entries)
		cache.Invalidate(path)
		_, ok := cache.Get(path)
		if ok {
			t.Error("Expected cache miss after Invalidate")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		cache.Put(path, entries)
		cache.Clear()
		_, ok := cache.Get(path)
		if ok {
			t.Error("Expected cache miss after Clear")
		}
	})
}
