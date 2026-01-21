package ops

import (
	"context"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
)

// CreateAtomic handles atomic creation of a file or directory with conflict resolution.
// It returns the final path used (which might be different from the requested path if Rename policy was used).
func CreateAtomic(ctx context.Context, fs core.FileSystem, path string, isDir bool, policy conflict.Policy) (string, error) {
	resolver := conflict.NewResolver()
	resolvedPath, _, err := resolver.Resolve(ctx, fs, "", path, policy)
	if err != nil {
		return "", err
	}

	if resolvedPath == "" {
		return "", nil // Skip
	}

	// If Overwrite policy and destination exists, remove it first
	if resolvedPath == path && policy == conflict.Overwrite {
		if err := fs.RemoveAll(ctx, path); err != nil {
			return "", err
		}
	}

	if isDir {
		if err := fs.MkdirAll(ctx, resolvedPath, 0755); err != nil {
			return "", err
		}
	} else {
		wc, err := fs.Create(ctx, resolvedPath)
		if err != nil {
			return "", err
		}
		wc.Close()
	}

	return resolvedPath, nil
}
