package core

import (
	"os"
	"testing"
	"time"
)

type mockFileInfo struct {
	os.FileInfo
	name string
}

func (m mockFileInfo) Name() string { return m.name }

func TestMetadataCache(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := NewMetadataCache(ttl)
	path := "/test/path"
	entries := []os.FileInfo{mockFileInfo{name: "file1"}}

	// Test Put and Get
	cache.Put(path, entries)
	got, ok := cache.Get(path)
	if !ok || len(got) != 1 || got[0].Name() != "file1" {
		t.Error("Cache Get failed after Put")
	}

	// Test Invalidate
	cache.Invalidate(path)
	_, ok = cache.Get(path)
	if ok {
		t.Error("Cache should be empty after Invalidate")
	}

	// Test Expiry
	cache.Put(path, entries)
	time.Sleep(ttl + 10*time.Millisecond)
	_, ok = cache.Get(path)
	if ok {
		t.Error("Cache should be expired")
	}

	// Test Clear
	cache.Put(path, entries)
	cache.Clear()
	_, ok = cache.Get(path)
	if ok {
		t.Error("Cache should be empty after Clear")
	}
}
