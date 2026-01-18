package ops

import (
	"context"
	"fmt"
	"io"
	"sync"

	"fm/internal/constants"
	"fm/internal/files/conflict"
	"fm/internal/files/core"
	"fm/internal/files/errors"

	"golang.org/x/sync/errgroup"
)

// Copy copies a file or directory recursively from src to dst within the same filesystem.
func Copy(ctx context.Context, fs core.FileSystem, src, dst string, progChan chan<- core.Progress, policy conflict.Policy) error {
	return CrossCopy(ctx, fs, fs, src, dst, progChan, policy)
}

// CrossCopy copies a file or directory between different filesystems.
func CrossCopy(ctx context.Context, srcFS, dstFS core.FileSystem, src, dst string, progChan chan<- core.Progress, policy conflict.Policy) error {
	if src == "" {
		return errors.WrapErrorWithPath(fmt.Errorf("no files selected"), "Copy", "")
	}

	if srcFS == dstFS {
		sAbs, _ := srcFS.Abs(src)
		dAbs, _ := dstFS.Abs(dst)
		if sAbs == dAbs {
			return errors.WrapErrorWithPath(fmt.Errorf("source and destination are the same"), "CrossCopy", src)
		}
	}

	// Resolve conflict if any
	resolver := conflict.NewResolver()
	resolvedPath, isRenamed, err := resolver.Resolve(ctx, dstFS, src, dst, policy)
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
	if resolvedPath == dst && policy == conflict.Overwrite {
		_ = dstFS.RemoveAll(ctx, dst)
	}
	dst = resolvedPath

	if progChan != nil && isRenamed {
		select {
		case progChan <- core.Progress{
			Percent: 0,
			Label:   fmt.Sprintf("Copying %s as %s...", srcFS.Base(src), dstFS.Base(dst)),
		}:
		default:
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	info, err := srcFS.Lstat(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "CrossCopy", src)
	}

	if info.IsDir() {
		return errors.WrapErrorWithPath(crossCopyDir(ctx, srcFS, dstFS, src, dst, progChan), "CrossCopy", fmt.Sprintf("%s -> %s", src, dst))
	}
	return errors.WrapErrorWithPath(crossCopyFile(ctx, srcFS, dstFS, src, dst, progChan), "CrossCopy", fmt.Sprintf("%s -> %s", src, dst))
}

func crossCopyFile(ctx context.Context, srcFS, dstFS core.FileSystem, src, dst string, progChan chan<- core.Progress) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	out, err := dstFS.Create(ctx, dst)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Create", dst)
	}
	defer out.Close()

	info, err := srcFS.Stat(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", src)
	}

	// Pre-allocate disk space if destination supports it
	if !info.IsDir() {
		_ = dstFS.Preallocate(ctx, dst, info.Size())
	}

	in, err := srcFS.Open(ctx, src)
	if err != nil {
		out.Close()
		_ = dstFS.RemoveAll(ctx, dst) // Clean up partial file
		return errors.WrapErrorWithPath(err, "Open", src)
	}
	defer in.Close()

	// Wrap with cancellable I/O
	cin := NewCancellableReader(ctx, in)
	cout := NewCancellableWriter(ctx, out)

	// Acquire buffer from pool for zero-allocation I/O
	buf := GetBuffer()
	defer PutBuffer(buf)

	if progChan != nil {
		pw := &progressWriter{
			Writer:   cout,
			Total:    info.Size(),
			Label:    "Copying " + srcFS.Base(src),
			progChan: progChan,
		}
		_, err = io.CopyBuffer(pw, cin, buf)
	} else {
		_, err = io.CopyBuffer(cout, cin, buf)
	}

	if err != nil {
		out.Close()
		_ = dstFS.RemoveAll(ctx, dst) // Clean up partial file
		return errors.WrapErrorWithPath(err, "CrossCopyFile", fmt.Sprintf("%s -> %s", src, dst))
	}

	return errors.WrapErrorWithPath(dstFS.Chmod(ctx, dst, info.Mode()), "Chmod", dst)
}

func crossCopyDir(ctx context.Context, srcFS, dstFS core.FileSystem, src, dst string, progChan chan<- core.Progress) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(constants.MaxCopyWorkers)
	var mu sync.Mutex
	err := crossCopyDirRecursive(ctx, srcFS, dstFS, src, dst, make(map[string]bool), &mu, progChan, g)
	if err != nil {
		return err
	}
	return g.Wait()
}

func crossCopyDirRecursive(ctx context.Context, srcFS, dstFS core.FileSystem, src, dst string, visited map[string]bool, mu *sync.Mutex, progChan chan<- core.Progress, g *errgroup.Group) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// We use srcFS for Abs check because we're visiting source items
	absSrc, err := srcFS.Abs(src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Abs", src)
	}

	mu.Lock()
	if visited[absSrc] {
		mu.Unlock()
		return nil // Skip already visited path to avoid loops
	}
	visited[absSrc] = true
	mu.Unlock()

	info, err := srcFS.Stat(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", src)
	}

	if err := dstFS.MkdirAll(ctx, dst, info.Mode()); err != nil {
		return errors.WrapErrorWithPath(err, "MkdirAll", dst)
	}

	entries, err := srcFS.ReadDir(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "ReadDir", src)
	}

	for _, entry := range entries {
		srcPath := srcFS.Join(src, entry.Name())
		dstPath := dstFS.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := crossCopyDirRecursive(ctx, srcFS, dstFS, srcPath, dstPath, visited, mu, progChan, g); err != nil {
				return err
			}
		} else {
			srcPath := srcPath // Capture for closure
			dstPath := dstPath // Capture for closure
			g.Go(func() error {
				return crossCopyFile(ctx, srcFS, dstFS, srcPath, dstPath, progChan)
			})
		}
	}
	return nil
}

// Deprecated: use CrossCopy instead for same FS
func copyFile(ctx context.Context, fs core.FileSystem, src, dst string, progChan chan<- core.Progress) error {
	return crossCopyFile(ctx, fs, fs, src, dst, progChan)
}

// Deprecated: use CrossCopy instead for same FS
func copyDir(ctx context.Context, fs core.FileSystem, src, dst string, progChan chan<- core.Progress) error {
	return crossCopyDir(ctx, fs, fs, src, dst, progChan)
}
