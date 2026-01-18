package git

import (
	"context"
	"os/exec"
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
)

func (gs *gitService) GetRoot(ctx context.Context, path string) string {
	if !gs.IsEnabled() {
		return ""
	}

	// Check cache
	if root, ok := gs.rootCache.Load(path); ok {
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
	gs.rootCache.Store(path, root)
	return root
}
