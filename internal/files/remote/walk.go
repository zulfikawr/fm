package remote

import (
	"context"
	"fmt"
	"path"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

func (fs *RemoteFS) Walk(ctx context.Context, root string, walkFn filepath.WalkFunc) error {
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

	state := &walkState{
		ctx:    ctx,
		g:      g,
		walkFn: walkFn,
	}

	if err := fs.parallelWalk(state, root); err != nil {
		return err
	}

	return g.Wait()
}

type walkState struct {
	ctx    context.Context
	g      *errgroup.Group
	walkFn filepath.WalkFunc
}

func (fs *RemoteFS) parallelWalk(state *walkState, root string) error {
	// Read current directory entries
	entries, err := fs.ReadDir(state.ctx, root)
	if err != nil {
		return state.walkFn(root, nil, err)
	}

	for i := range entries {
		entry := entries[i]
		p := path.Join(root, entry.Name())

		// Report the entry
		if err := state.walkFn(p, entry, nil); err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}

		// Recursively walk subdirectories in parallel
		if entry.IsDir() {
			state.g.Go(func() error {
				select {
				case <-state.ctx.Done():
					return state.ctx.Err()
				case <-fs.ctx.Done():
					return fmt.Errorf("filesystem closed")
				default:
					return fs.parallelWalk(state, p)
				}
			})
		}
	}

	return nil
}
