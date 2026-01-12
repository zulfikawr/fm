package git

import (
	"context"
	"pgregory.net/rapid"
	"testing"
)

func TestGitRootCaching_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property 8: Git Root Caching
	rapid.Check(t, func(t *rapid.T) {
		path := rapid.String().Draw(t, "path")
		expectedRoot := rapid.String().Draw(t, "root")

		s := NewGitService(true).(*gitService)
		s.rootCache.Store(path, expectedRoot)

		// The second call should return the cached value and not execute any command
		// (We verify this by ensuring the cached value is returned even if it's "fake")
		root := s.GetRoot(context.Background(), path)
		if root != expectedRoot {
			t.Errorf("Expected cached root %s, got %s", expectedRoot, root)
		}
	})
}

func TestGitDisabledBehavior_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property 9: Git Disabled Behavior
	rapid.Check(t, func(t *rapid.T) {
		path := rapid.String().Draw(t, "path")

		s := NewGitService(false)

		root := s.GetRoot(context.Background(), path)
		if root != "" {
			t.Errorf("Expected empty root when git is disabled, got %s", root)
		}

		statuses, branch := s.GetStatus(context.Background(), path)
		if statuses != nil || branch != "" {
			t.Errorf("Expected nil statuses and empty branch when git is disabled, got %v and %s", statuses, branch)
		}
	})
}
