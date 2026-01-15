package git

import (
	"fm/internal/testutil"
	"testing"
)

func TestGitService_Enabled(t *testing.T) {
	s := NewGitService(true)
	testutil.AssertEqual(t, true, s.IsEnabled(), "Should be enabled initially")

	s.SetEnabled(false)
	testutil.AssertEqual(t, false, s.IsEnabled(), "Should be disabled after SetEnabled(false)")
}
