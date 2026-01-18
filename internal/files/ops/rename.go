package ops

import (
	"context"
	"fmt"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
)

// Rename renames a file or directory.
func Rename(ctx context.Context, fs core.FileSystem, oldPath, newPath string, policy conflict.Policy) error {
	if oldPath == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("no files selected"), "Rename", "")
	}
	// Validate new filename component
	if err := ValidateFileName(fs.Base(newPath)); err != nil {
		return errors.WrapErrorWithPath(err, "Rename", newPath)
	}

	resolver := conflict.NewResolver()
	resolvedPath, _, err := resolver.Resolve(ctx, fs, oldPath, newPath, policy)
	if err != nil {
		if cerr, ok := err.(*conflict.ConflictError); ok {
			cerr.OpType = "rename"
			return cerr
		}
		return err
	}

	if resolvedPath == "" {
		return nil // Skip
	}
	newPath = resolvedPath

	return errors.WrapErrorWithPath(fs.Rename(ctx, oldPath, newPath), "Rename", fmt.Sprintf("%s -> %s", oldPath, newPath))
}
