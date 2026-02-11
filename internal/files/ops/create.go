package ops

import (
	"io"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/logger"
)

// CreateAtomic creates a new file or directory atomically.
func CreateAtomic(opts CreateOptions) (string, error) {
	resolver := conflict.NewResolver()
	resolvedPath, _, err := resolver.Resolve(opts.OpCtx.Context, opts.OpCtx.FS, conflict.ResolveOptions{
		Src:    "",
		Dst:    opts.Path,
		Policy: opts.Conflict.Policy,
	})
	if err != nil {
		return "", err
	}

	if resolvedPath == "" {
		return "", nil // Cancelled
	}

	if opts.IsDir {
		err = opts.OpCtx.FS.MkdirAll(opts.OpCtx.Context, resolvedPath, 0755)
	} else {
		var f io.WriteCloser
		f, err = opts.OpCtx.FS.Create(opts.OpCtx.Context, resolvedPath)
		if err == nil {
			logger.CloseAndLog(f, "newly created file")
		}
	}

	return resolvedPath, err
}
