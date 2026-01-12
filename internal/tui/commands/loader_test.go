package commands

import (
	"context"
	"fm/internal/testutil"
	"testing"

	"fm/internal/files/local"
	"fm/internal/files/sorting"
	"fm/internal/git"
)

type MockGitService struct {
	git.GitService
}

func (m *MockGitService) IsEnabled() bool {
	return true
}

func (m *MockGitService) GetStatus(ctx context.Context, path string) (map[string]string, string) {
	return make(map[string]string), "main"
}

func (m *MockGitService) GetRoot(ctx context.Context, path string) string {
	return path
}

func TestLoader(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	fs := &local.LocalFS{}
	gs := &MockGitService{}

	t.Run("Reload Command", func(t *testing.T) {
		cmd := Reload(fs, gs, tmpDir, 1, sorting.SortDefault, true)
		msg := cmd()
		if _, ok := msg.(LoadedItemsMsg); !ok {
			t.Errorf("Expected LoadedItemsMsg, got %T", msg)
		}
	})

	t.Run("FetchGitStatus Command", func(t *testing.T) {
		cmd := FetchGitStatus(gs, tmpDir)
		msg := cmd()
		if _, ok := msg.(GitStatusMsg); !ok {
			t.Errorf("Expected GitStatusMsg, got %T", msg)
		}
	})
}
