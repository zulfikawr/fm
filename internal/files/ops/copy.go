package ops

import (
	"fmt"
	"io"
	"sync"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/logger"

	"golang.org/x/sync/errgroup"
)

// Copy copies a file or directory recursively from src to dst within the same filesystem.
func Copy(opts CopyOptions) error {
	opts.SrcFS = opts.OpCtx.FS
	return CrossCopy(opts)
}

// CrossCopy copies a file or directory between different filesystems.
func CrossCopy(opts CopyOptions) error {
	if opts.Src == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("no files selected"), "Copy", "")
	}

	if opts.SrcFS == opts.OpCtx.FS {
		sAbs, _ := opts.SrcFS.Abs(opts.Src)
		dAbs, _ := opts.OpCtx.FS.Abs(opts.Dst)
		if sAbs == dAbs {
			return errors.WrapErrorWithPath(fmt.Errorf("source and destination are the same"), "CrossCopy", opts.Src)
		}
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
			cerr.IsMove = false
			cerr.OpType = "copy"
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
	opts.Dst = resolvedPath

	if opts.OpCtx.Progress != nil && isRenamed {
		select {
		case opts.OpCtx.Progress <- core.Progress{
			Percent: 0,
			Label:   fmt.Sprintf("Copying %s as %s...", opts.SrcFS.Base(opts.Src), opts.OpCtx.FS.Base(opts.Dst)),
		}:
		default:
		}
	}

	select {
	case <-opts.OpCtx.Context.Done():
		return opts.OpCtx.Context.Err()
	default:
	}

	info, err := opts.SrcFS.Lstat(opts.OpCtx.Context, opts.Src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "CrossCopy", opts.Src)
	}

	if info.IsDir() {
		return errors.WrapErrorWithPath(crossCopyDir(opts), "CrossCopy", fmt.Sprintf("%s -> %s", opts.Src, opts.Dst))
	}
	return errors.WrapErrorWithPath(crossCopyFile(opts), "CrossCopy", fmt.Sprintf("%s -> %s", opts.Src, opts.Dst))
}

func crossCopyFile(opts CopyOptions) error {
	select {
	case <-opts.OpCtx.Context.Done():
		return opts.OpCtx.Context.Err()
	default:
	}

	out, err := opts.OpCtx.FS.Create(opts.OpCtx.Context, opts.Dst)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Create", opts.Dst)
	}
	defer out.Close()

	info, err := opts.SrcFS.Stat(opts.OpCtx.Context, opts.Src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", opts.Src)
	}

	// Pre-allocate disk space if destination supports it
	if !info.IsDir() {
		if err := opts.OpCtx.FS.Preallocate(opts.OpCtx.Context, opts.Dst, info.Size()); err != nil {
			logger.Debugf("Preallocate not supported or failed: %v", err)
		}
	}

	in, err := opts.SrcFS.Open(opts.OpCtx.Context, opts.Src)
	if err != nil {
		out.Close()
		if err := opts.OpCtx.FS.RemoveAll(opts.OpCtx.Context, opts.Dst); err != nil {
			logger.Warnf("Failed to clean up partial file %s: %v", opts.Dst, err)
		}
		return errors.WrapErrorWithPath(err, "Open", opts.Src)
	}
	defer in.Close()

	// Wrap with cancellable I/O
	cin := NewCancellableReader(opts.OpCtx.Context, in)
	cout := NewCancellableWriter(opts.OpCtx.Context, out)

	// Acquire buffer from pool for zero-allocation I/O
	buf := GetBuffer()
	defer PutBuffer(buf)

	if opts.OpCtx.Progress != nil {
		pw := &progressWriter{
			Writer:   cout,
			Total:    info.Size(),
			Label:    "Copying " + opts.SrcFS.Base(opts.Src),
			progChan: opts.OpCtx.Progress,
		}
		_, err = io.CopyBuffer(pw, cin, buf)
	} else {
		_, err = io.CopyBuffer(cout, cin, buf)
	}

	if err != nil {
		out.Close()
		if err := opts.OpCtx.FS.RemoveAll(opts.OpCtx.Context, opts.Dst); err != nil {
			logger.Warnf("Failed to clean up partial file %s: %v", opts.Dst, err)
		}
		return errors.WrapErrorWithPath(err, "CrossCopyFile", fmt.Sprintf("%s -> %s", opts.Src, opts.Dst))
	}

	return errors.WrapErrorWithPath(opts.OpCtx.FS.Chmod(opts.OpCtx.Context, opts.Dst, info.Mode()), "Chmod", opts.Dst)
}

func crossCopyDir(opts CopyOptions) error {
	g, _ := errgroup.WithContext(opts.OpCtx.Context)
	g.SetLimit(constants.MaxCopyWorkers)
	var mu sync.Mutex
	state := &copyState{
		opts:    opts,
		visited: make(map[string]bool),
		mu:      &mu,
		g:       g,
	}
	err := crossCopyDirRecursive(state, opts.Src, opts.Dst)
	if err != nil {
		return err
	}
	return g.Wait()
}

type copyState struct {
	opts    CopyOptions
	visited map[string]bool
	mu      *sync.Mutex
	g       *errgroup.Group
}

func crossCopyDirRecursive(state *copyState, src, dst string) error {
	select {
	case <-state.opts.OpCtx.Context.Done():
		return state.opts.OpCtx.Context.Err()
	default:
	}

	// We use srcFS for Abs check because we're visiting source items
	absSrc, err := state.opts.SrcFS.Abs(src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Abs", src)
	}

	state.mu.Lock()
	if state.visited[absSrc] {
		state.mu.Unlock()
		return nil // Skip already visited path to avoid loops
	}
	visited := state.visited
	visited[absSrc] = true
	state.mu.Unlock()

	info, err := state.opts.SrcFS.Stat(state.opts.OpCtx.Context, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", src)
	}

	if err := state.opts.OpCtx.FS.MkdirAll(state.opts.OpCtx.Context, dst, info.Mode()); err != nil {
		return errors.WrapErrorWithPath(err, "MkdirAll", dst)
	}

	entries, err := state.opts.SrcFS.ReadDir(state.opts.OpCtx.Context, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "ReadDir", src)
	}

	for _, entry := range entries {
		srcPath := state.opts.SrcFS.Join(src, entry.Name())
		dstPath := state.opts.OpCtx.FS.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := crossCopyDirRecursive(state, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			srcPath := srcPath // Capture for closure
			dstPath := dstPath // Capture for closure
			state.g.Go(func() error {
				opts := state.opts
				opts.Src = srcPath
				opts.Dst = dstPath
				return crossCopyFile(opts)
			})
		}
	}
	return nil
}
