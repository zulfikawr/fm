package cache

import "testing"

func TestSimpleCache(t *testing.T) {
	c := NewSimpleCache(2)

	t.Run("Put and Get", func(t *testing.T) {
		c.Put("a", 1)
		val, ok := c.Get("a")
		if !ok || val != 1 {
			t.Errorf("Expected 1, got %v", val)
		}
	})

	t.Run("Update existing", func(t *testing.T) {
		c.Put("a", 2)
		val, ok := c.Get("a")
		if !ok || val != 2 {
			t.Errorf("Expected 2, got %v", val)
		}
	})

	t.Run("Eviction", func(t *testing.T) {
		c.Put("b", 3)
		c.Put("c", 4) // Should evict "a" (if "a" was the oldest, which it is)

		_, ok := c.Get("a")
		if ok {
			t.Error("Expected 'a' to be evicted")
		}

		val, ok := c.Get("b")
		if !ok || val != 3 {
			t.Errorf("Expected 'b' to be 3, got %v", val)
		}

		val, ok = c.Get("c")
		if !ok || val != 4 {
			t.Errorf("Expected 'c' to be 4, got %v", val)
		}
	})
}
