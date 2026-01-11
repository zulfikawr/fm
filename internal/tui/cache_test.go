package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLRUCache(t *testing.T) {
	c := NewLRUCache(2)

	now := time.Now()
	e1 := SizeCacheEntry{Size: 100, MTime: now}
	e2 := SizeCacheEntry{Size: 200, MTime: now}
	e3 := SizeCacheEntry{Size: 300, MTime: now}

	c.Put("key1", e1)
	c.Put("key2", e2)

	if val, ok := c.Get("key1"); !ok || val.Size != 100 {
		t.Errorf("Expected key1 with size 100, got %v, %v", val.Size, ok)
	}

	// key1 should now be most recent, key2 oldest
	c.Put("key3", e3) // should evict key2

	if _, ok := c.Get("key2"); ok {
		t.Error("key2 should have been evicted")
	}

	if val, ok := c.Get("key1"); !ok || val.Size != 100 {
		t.Errorf("Expected key1 with size 100, got %v, %v", val.Size, ok)
	}

	// Update existing
	e1.Size = 101
	c.Put("key1", e1)
	if val, ok := c.Get("key1"); !ok || val.Size != 101 {
		t.Errorf("Expected key1 with size 101, got %v, %v", val.Size, ok)
	}
}

func TestLRUCachePersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fm-lru-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cacheFile := filepath.Join(tmpDir, "cache.gob")
	c := NewLRUCache(10)

	now := time.Now().Truncate(time.Second) // Gob might have precision differences
	e1 := SizeCacheEntry{Size: 100, MTime: now}
	c.Put("key1", e1)

	if err := c.Save(cacheFile); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	c2 := NewLRUCache(10)
	if err := c2.Load(cacheFile); err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if val, ok := c2.Get("key1"); !ok || val.Size != 100 {
		t.Errorf("Expected key1 with size 100, got %v, %v", val.Size, ok)
	}

	// Test missing file
	if err := c2.Load("non-existent"); err != nil {
		t.Errorf("Loading missing file should not return error, got %v", err)
	}
}

func TestSimpleCache(t *testing.T) {
	c := NewSimpleCache(2)

	c.Put("k1", 1)
	c.Put("k2", 2)

	if val, ok := c.Get("k1"); !ok || val != 1 {
		t.Errorf("Expected k1=1, got %v, %v", val, ok)
	}

	c.Put("k3", 3) // Should evict k1 (oldest)

	if _, ok := c.Get("k1"); ok {
		t.Error("k1 should have been evicted")
	}

	if val, ok := c.Get("k2"); !ok || val != 2 {
		t.Errorf("Expected k2=2, got %v, %v", val, ok)
	}

	// Update existing
	c.Put("k1", 10)
	if val, ok := c.Get("k1"); !ok || val != 10 {
		t.Errorf("Expected k1=10, got %v, %v", val, ok)
	}
}
