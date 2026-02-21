package ops

import (
	"context"
	"fmt"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/files/trash"
	"github.com/zulfikawr/fm/internal/logger"
)

// Delete removes a file or directory recursively.
func Delete(opts DeleteOptions) (err error) {
	// Recover from panics and convert to errors
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during delete: %v", r)
			logger.Errorf("Delete panic recovered: %v", r)
		}
	}()

	if len(opts.Paths) == 0 || opts.Paths[0] == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("no files selected"), "Delete", "")
	}
	path := opts.Paths[0]
	select {
	case <-opts.OpCtx.Context.Done():
		return opts.OpCtx.Context.Err()
	default:
	}

	if opts.OpCtx.Progress != nil {
		select {
		case opts.OpCtx.Progress <- core.Progress{
			Percent: 0,
			Label:   "Deleting " + opts.OpCtx.FS.Base(path) + "...",
		}:
		default:
		}
	}

	// Use trash if enabled
	if opts.Trash.UseTrash {
		if err := MoveToTrash(opts.OpCtx.Context, opts.OpCtx.FS, path); err != nil {
			return errors.WrapErrorWithPath(err, "Delete", path)
		}
	} else {
		err := opts.OpCtx.FS.RemoveAll(opts.OpCtx.Context, path)
		if err != nil {
			return errors.WrapErrorWithPath(err, "Delete", path)
		}
	}

	if opts.OpCtx.Progress != nil {
		select {
		case opts.OpCtx.Progress <- core.Progress{
			Percent: 1.0,
			Label:   "Deleted " + opts.OpCtx.FS.Base(path),
		}:
		default:
		}
	}

	return nil
}

// MoveToTrash moves a file or directory to the custom trash.
func MoveToTrash(ctx context.Context, fs core.FileSystem, path string) error {
	if path == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("no files selected"), "Trash", "")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if !fs.IsLocal() {
		return errors.WrapErrorWithPath(&errors.UnsupportedOperationError{Op: "Trash", Filesystem: "Remote"}, "Trash", path)
	}

	manager, err := trash.NewManager(fs)
	if err != nil {
		return errors.WrapErrorWithPath(fmt.Errorf("create trash manager: %w", err), "Trash", path)
	}

	if err := manager.MoveToTrash(ctx, path); err != nil {
		return errors.WrapErrorWithPath(err, "Trash", path)
	}

	return nil
}
