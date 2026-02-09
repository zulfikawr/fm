package ops

import (
	"fmt"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
)

// DeleteMultiple removes multiple files or directories recursively.
func DeleteMultiple(opts DeleteOptions) error {
	if len(opts.Paths) > 0 {
		if err := ValidateWritable(opts.OpCtx.Context, opts.OpCtx.FS, core.GetParent(opts.OpCtx.FS, opts.Paths[0])); err != nil {
			return err
		}
	}
	for i, path := range opts.Paths {
		select {
		case <-opts.OpCtx.Context.Done():
			return opts.OpCtx.Context.Err()
		default:
		}

		if opts.OpCtx.Progress != nil && !opts.UseTrash {
			select {
			case opts.OpCtx.Progress <- core.Progress{
				Percent: float64(i) / float64(len(opts.Paths)),
				Label:   "Deleting " + opts.OpCtx.FS.Base(path) + "...",
			}:
			default:
			}
		}

		var err error
		if opts.UseTrash {
			err = MoveToTrash(opts.OpCtx.Context, opts.OpCtx.FS, path)
		} else {
			err = Delete(DeleteOptions{
				OpCtx: opts.OpCtx,
				Paths: []string{path},
			})
		}
		if err != nil {
			return err
		}
	}

	if opts.OpCtx.Progress != nil && len(opts.Paths) > 0 {
		label := ""
		if len(opts.Paths) == 1 {
			label = fmt.Sprintf("Deleted %s", opts.OpCtx.FS.Base(opts.Paths[0]))
		} else {
			label = fmt.Sprintf("Deleted %d items", len(opts.Paths))
		}
		select {
		case opts.OpCtx.Progress <- core.Progress{
			Percent: 1.0,
			Label:   label,
		}:
		default:
		}
	}

	return nil
}

// CopyMultiple copies multiple items from sources to destDir between different filesystems.
func CopyMultiple(opts BatchOptions) error {
	if err := ValidateWritable(opts.OpCtx.Context, opts.OpCtx.FS, opts.DestDir); err != nil {
		return err
	}
	resolver := conflict.NewResolver()
	for i, src := range opts.Sources {
		select {
		case <-opts.OpCtx.Context.Done():
			return opts.OpCtx.Context.Err()
		default:
		}

		dst := opts.OpCtx.FS.Join(opts.DestDir, opts.SrcFS.Base(src))

		resolvedDst, isRenamed, err := resolver.Resolve(opts.OpCtx.Context, opts.OpCtx.FS, conflict.ResolveOptions{
			Src:    src,
			Dst:    dst,
			Policy: opts.Conflict.Policy,
		})
		if err != nil {
			if cerr, ok := err.(*conflict.ConflictError); ok {
				cerr.PendingItems = opts.Sources[i:]
				cerr.OpType = "copy"
				return cerr
			}
			return err
		}

		if resolvedDst == "" {
			// If we skipped and not applying to all, reset policy for next items
			if !opts.Conflict.ApplyToAll {
				opts.Conflict.Policy = conflict.Ask
			}
			continue // Skip
		}
		dst = resolvedDst

		if opts.OpCtx.Progress != nil {
			label := "Copying " + opts.SrcFS.Base(src) + "..."
			if isRenamed {
				label = fmt.Sprintf("Copying %s as %s...", opts.SrcFS.Base(src), opts.OpCtx.FS.Base(dst))
			}
			select {
			case opts.OpCtx.Progress <- core.Progress{
				Percent: float64(i) / float64(len(opts.Sources)),
				Label:   label,
			}:
			default:
			}
		}

		copyOpts := CopyOptions{
			OpCtx: opts.OpCtx,
			SrcFS: opts.SrcFS,
			Src:   src,
			Dst:   dst,
			Conflict: ConflictOptions{
				Policy: conflict.Overwrite,
			},
		}
		if err := CrossCopy(copyOpts); err != nil {
			return err
		}

		// If we are not applying to all, reset policy to Ask after the first successful resolution
		if !opts.Conflict.ApplyToAll {
			opts.Conflict.Policy = conflict.Ask
		}
	}
	return nil
}

// MoveMultiple moves multiple items from sources to destDir between different filesystems.
func MoveMultiple(opts BatchOptions) error {
	if err := ValidateWritable(opts.OpCtx.Context, opts.OpCtx.FS, opts.DestDir); err != nil {
		return err
	}
	resolver := conflict.NewResolver()
	for i, src := range opts.Sources {
		select {
		case <-opts.OpCtx.Context.Done():
			return opts.OpCtx.Context.Err()
		default:
		}

		dst := opts.OpCtx.FS.Join(opts.DestDir, opts.SrcFS.Base(src))

		resolvedDst, isRenamed, err := resolver.Resolve(opts.OpCtx.Context, opts.OpCtx.FS, conflict.ResolveOptions{
			Src:    src,
			Dst:    dst,
			Policy: opts.Conflict.Policy,
		})
		if err != nil {
			if cerr, ok := err.(*conflict.ConflictError); ok {
				cerr.PendingItems = opts.Sources[i:]
				cerr.IsMove = true
				cerr.OpType = "move"
				return cerr
			}
			return err
		}

		if resolvedDst == "" {
			if !opts.Conflict.ApplyToAll {
				opts.Conflict.Policy = conflict.Ask
			}
			continue // Skip
		}
		dst = resolvedDst

		if opts.OpCtx.Progress != nil {
			label := "Moving " + opts.SrcFS.Base(src) + "..."
			if isRenamed {
				label = fmt.Sprintf("Moving %s as %s...", opts.SrcFS.Base(src), opts.OpCtx.FS.Base(dst))
			}
			select {
			case opts.OpCtx.Progress <- core.Progress{
				Percent: float64(i) / float64(len(opts.Sources)),
				Label:   label,
			}:
			default:
			}
		}

		moveOpts := CopyOptions{
			OpCtx: opts.OpCtx,
			SrcFS: opts.SrcFS,
			Src:   src,
			Dst:   dst,
			Conflict: ConflictOptions{
				Policy: conflict.Overwrite,
			},
		}
		if err := CrossMove(moveOpts); err != nil {
			return err
		}

		if !opts.Conflict.ApplyToAll {
			opts.Conflict.Policy = conflict.Ask
		}
	}
	return nil
}

// CheckAndMarkProcessing checks if any of the paths are currently being processed
// and marks them as processing if none are. Returns false if any path is already processing.
func CheckAndMarkProcessing(processing map[string]bool, paths []string) bool {
	for _, p := range paths {
		if processing[p] {
			return false
		}
	}
	for _, p := range paths {
		processing[p] = true
	}
	return true
}
