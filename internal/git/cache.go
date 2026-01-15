package git

import (
	"context"
	"os/exec"
	"strings"

	"fm/internal/constants"
)

func (s *gitService) GetRoot(ctx context.Context, path string) string {
	if !s.IsEnabled() {
		return ""
	}

	// Check cache
	if root, ok := s.rootCache.Load(path); ok {
		return root.(string)
	}

	ctx, cancel := context.WithTimeout(ctx, constants.GitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	root := strings.TrimSpace(string(out))
	s.rootCache.Store(path, root)
	return root
}
