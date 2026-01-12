package git

import (
	"context"
	"pgregory.net/rapid"
	"testing"
)

func TestGetRootCaching_Detailed_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property 8: Git Root Caching
	rapid.Check(t, func(t *rapid.T) {
		path := rapid.String().Draw(t, "path")
		root := rapid.String().Draw(t, "root")

		s := NewGitService(true).(*gitService)

		// Manually prime the cache
		s.rootCache.Store(path, root)

		// Verify GetRoot returns the cached value
		result := s.GetRoot(context.Background(), path)
		if result != root {
			t.Errorf("Expected root %s from cache, got %s", root, result)
		}
	})
}
