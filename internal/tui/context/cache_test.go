package context_test

import (
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestSimpleCache(t *testing.T) {
	c := core.NewSimpleCache[string, int](2, 0)

	t.Run("Put and Get", func(t *testing.T) {
		c.Put("a", 1)
		val, ok := c.Get("a")
		testutil.AssertEqual(t, true, ok, "Should find 'a'")
		testutil.AssertEqual(t, 1, val, "Value should match")
	})

	t.Run("Update", func(t *testing.T) {
		c.Put("a", 2)
		val, ok := c.Get("a")
		testutil.AssertEqual(t, true, ok, "Should find 'a'")
		testutil.AssertEqual(t, 2, val, "Value should be updated")
	})

	t.Run("Eviction", func(t *testing.T) {
		c.Put("b", 10)
		c.Put("c", 20) // Should evict 'a'

		_, ok := c.Get("a")
		testutil.AssertEqual(t, false, ok, "Should have evicted 'a'")

		val, _ := c.Get("b")
		testutil.AssertEqual(t, 10, val, "Should still have 'b'")

		val, _ = c.Get("c")
		testutil.AssertEqual(t, 20, val, "Should have 'c'")
	})
}
