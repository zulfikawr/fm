package git

import (
	"context"
	"sync"
)

// GitService provides git repository operations
type GitService interface {
	// GetStatus returns file statuses and branch for a directory
	GetStatus(ctx context.Context, path string) (statuses map[string]string, branch string)

	// GetRoot returns the git repository root for a path, cached
	GetRoot(ctx context.Context, path string) string

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

func (s *gitService) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

func (s *gitService) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}
