package git

import (
	"context"
	"testing"

	"fm/internal/testutil"
)

func TestGitService_GetRoot(t *testing.T) {
	ctx := context.Background()
	gs := NewGitService(true)

	t.Run("Disabled service", func(t *testing.T) {
		gs.SetEnabled(false)
		root := gs.GetRoot(ctx, "/any/path")
		testutil.AssertEqual(t, "", root, "Should return empty if disabled")
	})

	t.Run("Enabled but no repo", func(t *testing.T) {
		gs.SetEnabled(true)
		// This will actually run 'git rev-parse' which might fail in /tmp or similar
		// But it won't crash.
	})
}
