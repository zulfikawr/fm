package git

import (
	"context"
	"sync"
)

// GitService provides git repository operations
type GitService interface {
	// GetStatus returns file statuses and branch for a directory
	GetStatus(ctx context.Context, path string) (statuses map[string]string, branch string, modified, staged, untracked int)

	// GetRoot returns the git repository root for a path, cached
	GetRoot(ctx context.Context, path string) string

	// GetIgnoredFiles returns a list of ignored files in the repository
	GetIgnoredFiles(ctx context.Context, repoRoot string) ([]string, error)

	// IsEnabled returns whether git integration is enabled
	IsEnabled() bool

	// SetEnabled enables or disables git integration
	SetEnabled(enabled bool)
}

// gitService implements GitService
type gitService struct {
	enabled   bool
	rootCache sync.Map // path -> root mapping
	mu        sync.RWMutex
}

// NewGitService creates a new GitService
func NewGitService(enabled bool) GitService {
	return &gitService{
		enabled: enabled,
	}
}

func (gs *gitService) IsEnabled() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.enabled
}

func (gs *gitService) SetEnabled(enabled bool) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.enabled = enabled
}
