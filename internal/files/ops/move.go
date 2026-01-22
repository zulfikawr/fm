package ops

import (
	"fmt"
	"time"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/logger"
)

// Move moves a file or directory. It tries Rename first, and falls back to Copy+Delete if Rename fails.
func Move(opts CopyOptions) error {
	newOpts := opts
	newOpts.SrcFS = opts.OpCtx.FS
	return CrossMove(newOpts)
}

// CrossMove moves a file or directory between different filesystems.
func CrossMove(opts CopyOptions) error {
	if opts.Src == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("no files selected"), "Move", "")
	}

	// Resolve conflict if any
	resolver := conflict.NewResolver()
	resolvedPath, isRenamed, err := resolver.Resolve(opts.OpCtx.Context, opts.OpCtx.FS, conflict.ResolveOptions{
		Src:    opts.Src,
		Dst:    opts.Dst,
		Policy: opts.Conflict.Policy,
	})
	if err != nil {
		if cerr, ok := err.(*conflict.ConflictError); ok {
			cerr.IsMove = true
			cerr.OpType = "move"
			return cerr
		}
		return err
	}

	if resolvedPath == "" {
		return nil // Skip
	}

	if resolvedPath == opts.Dst && opts.Conflict.Policy == conflict.Overwrite {
		if err := opts.OpCtx.FS.RemoveAll(opts.OpCtx.Context, opts.Dst); err != nil {
			logger.Warnf("Failed to remove existing item for overwrite: %v", err)
		}
	}

	if opts.OpCtx.Progress != nil && isRenamed {
		select {
		case opts.OpCtx.Progress <- core.Progress{
			Percent: 0,
			Label:   fmt.Sprintf("Moving %s as %s...", opts.SrcFS.Base(opts.Src), opts.OpCtx.FS.Base(resolvedPath)),
		}:
		default:
		}
	}

	// 1. Try atomic rename first if same FS
	if opts.SrcFS == opts.OpCtx.FS {
		err := opts.SrcFS.Rename(opts.OpCtx.Context, opts.Src, resolvedPath)
		if err == nil {
			if opts.OpCtx.Progress != nil {
				select {
				case opts.OpCtx.Progress <- core.Progress{Percent: 1.0, Label: "Moved " + opts.SrcFS.Base(opts.Src)}:
				default:
				}
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		}
	}

	// Create a new options object for the copy-fallback with the resolved destination
	copyOpts := opts
	copyOpts.Dst = resolvedPath
	copyOpts.Conflict.Policy = conflict.Overwrite

	// 2. Fallback for cross-device/FS moves: Copy then Delete
	if err := CrossCopy(copyOpts); err != nil {
		if cleanupErr := opts.OpCtx.FS.RemoveAll(opts.OpCtx.Context, resolvedPath); cleanupErr != nil {
			logger.Warnf("Failed to clean up destination after failed move: %v", cleanupErr)
		}
		return err
	}

	// 3. Verify the copy was successful using the new destination
	verifyOpts := opts
	verifyOpts.Dst = resolvedPath
	if err := verifyCrossMove(verifyOpts); err != nil {
		if cleanupErr := opts.OpCtx.FS.RemoveAll(opts.OpCtx.Context, resolvedPath); cleanupErr != nil {
			logger.Warnf("Failed to clean up destination after failed move verification: %v", cleanupErr)
		}
		return errors.WrapErrorWithPath(fmt.Errorf("move verification failed: %w", err), "CrossMove", opts.Src)
	}

	// 4. "Commit" the move by deleting the source
	deleteOpts := DeleteOptions{
		OpCtx: OpContext{
			Context:  opts.OpCtx.Context,
			FS:       opts.SrcFS,
			Progress: nil,
		},
		Paths: []string{opts.Src},
	}
	if err := Delete(deleteOpts); err != nil {
		return errors.WrapErrorWithPath(fmt.Errorf("move partially successful: items copied and verified but failed to remove from source: %w", err), "CrossMove", opts.Src)
	}

	return nil
}

// verifyCrossMove performs a basic check to ensure the source was copied correctly
func verifyCrossMove(opts CopyOptions) error {
	sInfo, err := opts.SrcFS.Lstat(opts.OpCtx.Context, opts.Src)
	if err != nil {
		return err
	}

	dInfo, err := opts.OpCtx.FS.Lstat(opts.OpCtx.Context, opts.Dst)
	if err != nil {
		return err
	}

	if !sInfo.IsDir() && sInfo.Size() != dInfo.Size() {
		return fmt.Errorf("size mismatch: source %d, destination %d", sInfo.Size(), dInfo.Size())
	}

	return nil
}
