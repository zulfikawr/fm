package remote

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

func (fs *RemoteFS) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(16) // Concurrency limit for parallel walking

	// Start with the root
	info, err := fs.Stat(ctx, root)
	if err := walkFn(root, info, err); err != nil {
		if err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if info == nil || !info.IsDir() {
		return nil
	}

	if err := fs.parallelWalk(ctx, g, root, walkFn); err != nil {
		return err
	}

	return g.Wait()
}

func (fs *RemoteFS) parallelWalk(ctx context.Context, g *errgroup.Group, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	// Read current directory entries
	entries, err := fs.ReadDir(ctx, root)
	if err != nil {
		return walkFn(root, nil, err)
	}

	for _, entry := range entries {
		entry := entry // capture
		p := path.Join(root, entry.Name())

		// Report the entry
		if err := walkFn(p, entry, nil); err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}

		// Recursively walk subdirectories in parallel
		if entry.IsDir() {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-fs.ctx.Done():
					return fmt.Errorf("filesystem closed")
				default:
					return fs.parallelWalk(ctx, g, p, walkFn)
				}
			})
		}
	}

	return nil
}
