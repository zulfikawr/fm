package git

import (
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestGitService_Enabled(t *testing.T) {
	gs := NewGitService(true)
	testutil.AssertEqual(t, true, gs.IsEnabled(), "Should be enabled initially")

	gs.SetEnabled(false)
	testutil.AssertEqual(t, false, gs.IsEnabled(), "Should be disabled after SetEnabled(false)")
}
