package git

import (
	"context"
	"fm/internal/testutil"
	"testing"
)

func TestGitService_GetRoot(t *testing.T) {
	ctx := context.Background()
	s := NewGitService(true)

	t.Run("Disabled service", func(t *testing.T) {
		s.SetEnabled(false)
		root := s.GetRoot(ctx, "/any/path")
		testutil.AssertEqual(t, "", root, "Should return empty if disabled")
	})

	t.Run("Enabled but no repo", func(t *testing.T) {
		s.SetEnabled(true)
		// This will actually run 'git rev-parse' which might fail in /tmp or similar
		// But it won't crash.
	})
}
