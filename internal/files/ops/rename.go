package ops

import (
	"fmt"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/errors"
)

// Rename renames a file or directory.
func Rename(opts RenameOptions) error {
	if opts.OldPath == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("source path is empty"), "Rename", "")
	}
	if err := ValidateFileName(opts.OpCtx.FS.Base(opts.NewPath)); err != nil {
		return errors.WrapErrorWithPath(err, "Rename", opts.NewPath)
	}

	resolver := conflict.NewResolver()
	resolvedPath, _, err := resolver.Resolve(opts.OpCtx.Context, opts.OpCtx.FS, conflict.ResolveOptions{
		Src:    opts.OldPath,
		Dst:    opts.NewPath,
		Policy: opts.Conflict.Policy,
	})
	if err != nil {
		return err
	}

	if resolvedPath == "" {
		return nil // Cancelled
	}

	return opts.OpCtx.FS.Rename(opts.OpCtx.Context, opts.OldPath, resolvedPath)
}
