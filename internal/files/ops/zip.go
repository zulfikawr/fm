package ops

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zulfikawr/fm/internal/files/conflict"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/logger"
)

// Zip compresses multiple files or directories into a single zip archive.
func Zip(opts ZipOptions) error {
	if len(opts.Srcs) == 0 {
		return errors.WrapError(fmt.Errorf("no source files specified"), "Zip")
	}

	resolver := conflict.NewResolver()
	resolvedDst, _, err := resolver.Resolve(opts.OpCtx.Context, opts.OpCtx.FS, conflict.ResolveOptions{
		Src:    opts.Srcs[0],
		Dst:    opts.Dst,
		Policy: opts.Conflict.Policy,
	})
	if err != nil {
		if cerr, ok := err.(*conflict.ConflictError); ok {
			cerr.PendingItems = opts.Srcs
			cerr.OpType = "zip"
			return cerr
		}
		return err
	}

	if resolvedDst == "" {
		return nil // Skip
	}

	// Create destination file
	out, err := opts.OpCtx.FS.Create(opts.OpCtx.Context, resolvedDst)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Create", resolvedDst)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	// For simplicity, we'll increment progress per file processed
	processedFiles := 0

	state := &zipState{
		opts: opts,
		zw:   zw,
		onFile: func() {
			processedFiles++
			if opts.OpCtx.Progress != nil {
				opts.OpCtx.Progress <- core.Progress{
					Label:   fmt.Sprintf("Zipping %d items into %s", len(opts.Srcs), opts.OpCtx.FS.Base(resolvedDst)),
					Percent: float64(processedFiles) / float64(len(opts.Srcs)), // Simplified progress
				}
			}
		},
	}

	for _, src := range opts.Srcs {
		select {
		case <-opts.OpCtx.Context.Done():
			return opts.OpCtx.Context.Err()
		default:
		}

		baseDir := opts.OpCtx.FS.Dir(src)

		err = walkAndZip(state, src, baseDir)
		if err != nil {
			return err
		}
	}

	return nil
}

type zipState struct {
	opts   ZipOptions
	zw     *zip.Writer
	onFile func()
}

func walkAndZip(state *zipState, currentPath, baseDir string) error {
	select {
	case <-state.opts.OpCtx.Context.Done():
		return state.opts.OpCtx.Context.Err()
	default:
	}

	info, err := state.opts.OpCtx.FS.Lstat(state.opts.OpCtx.Context, currentPath)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", currentPath)
	}

	// Calculate relative path for the zip header
	relPath, err := state.opts.OpCtx.FS.Rel(baseDir, currentPath)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Rel", currentPath)
	}
	// zip paths should always use forward slashes
	relPath = filepath.ToSlash(relPath)

	if info.IsDir() {
		// Ensure directory entry ends with a slash
		if !strings.HasSuffix(relPath, "/") {
			relPath += "/"
		}
		_, err = state.zw.Create(relPath)
		if err != nil {
			return errors.WrapErrorWithPath(err, "ZipCreateDir", relPath)
		}

		entries, err := state.opts.OpCtx.FS.ReadDir(state.opts.OpCtx.Context, currentPath)
		if err != nil {
			return errors.WrapErrorWithPath(err, "ReadDir", currentPath)
		}

		for _, entry := range entries {
			childPath := state.opts.OpCtx.FS.Join(currentPath, entry.Name())
			if err := walkAndZip(state, childPath, baseDir); err != nil {
				return err
			}
		}
		return nil
	}

	// File entry
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return errors.WrapErrorWithPath(err, "ZipHeader", currentPath)
	}
	header.Name = relPath
	header.Method = zip.Deflate

	writer, err := state.zw.CreateHeader(header)
	if err != nil {
		return errors.WrapErrorWithPath(err, "ZipCreate", currentPath)
	}

	in, err := state.opts.OpCtx.FS.Open(state.opts.OpCtx.Context, currentPath)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Open", currentPath)
	}
	defer in.Close()

	// Use a buffer for copying
	buf := GetBuffer()
	defer PutBuffer(buf)

	_, err = io.CopyBuffer(writer, in, buf)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Copy", currentPath)
	}

	state.onFile()
	return nil
}

// Unzip extracts a zip archive to the specified destination directory.
func Unzip(opts ZipOptions) error {
	resolver := conflict.NewResolver()
	resolvedDst, _, err := resolver.Resolve(opts.OpCtx.Context, opts.OpCtx.FS, conflict.ResolveOptions{
		Src:    opts.Src,
		Dst:    opts.Dst,
		Policy: opts.Conflict.Policy,
	})
	if err != nil {
		if cerr, ok := err.(*conflict.ConflictError); ok {
			cerr.PendingItems = []string{opts.Src}
			cerr.OpType = "unzip"
			return cerr
		}
		return err
	}

	if resolvedDst == "" {
		return nil // Skip
	}

	// Open the source file
	in, err := opts.OpCtx.FS.Open(opts.OpCtx.Context, opts.Src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Open", opts.Src)
	}
	defer in.Close()

	// zip.NewReader needs a ReaderAt and the size.
	// We'll get the size first.
	info, err := opts.OpCtx.FS.Stat(opts.OpCtx.Context, opts.Src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", opts.Src)
	}
	size := info.Size()

	// If it's a local file, we can use os.Open directly for ReaderAt
	var readerAt io.ReaderAt
	if opts.OpCtx.FS.IsLocal() {
		f, err := os.Open(opts.Src)
		if err != nil {
			return errors.WrapErrorWithPath(err, "OpenLocal", opts.Src)
		}
		defer f.Close()
		readerAt = f
	} else {
		// For remote files, we'll download to a temporary file first
		// because zip.NewReader requires random access (ReaderAt)
		tmpFile, err := os.CreateTemp("", "fm-unzip-*.zip")
		if err != nil {
			return errors.WrapError(err, "CreateTemp")
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = io.Copy(tmpFile, in)
		if err != nil {
			return errors.WrapError(err, "DownloadRemoteZip")
		}
		readerAt = tmpFile
	}

	zr, err := zip.NewReader(readerAt, size)
	if err != nil {
		return errors.WrapError(err, "NewZipReader")
	}

	totalFiles := len(zr.File)
	processedFiles := 0

	for _, f := range zr.File {
		select {
		case <-opts.OpCtx.Context.Done():
			return opts.OpCtx.Context.Err()
		default:
		}

		fpath, err := conflict.ValidateSecurePath(opts.OpCtx.FS, resolvedDst, f.Name)
		if err != nil {
			continue // Skip dangerous paths
		}

		if f.FileInfo().IsDir() {
			err = opts.OpCtx.FS.MkdirAll(opts.OpCtx.Context, fpath, f.Mode()|0111) // Ensure execute permission for directory access
			if err != nil {
				return errors.WrapErrorWithPath(err, "MkdirAll", fpath)
			}
		} else {
			// Check for conflict
			entryResolvedPath, isRenamed, err := resolver.Resolve(opts.OpCtx.Context, opts.OpCtx.FS, conflict.ResolveOptions{
				Src:    f.Name,
				Dst:    fpath,
				Policy: opts.Conflict.Policy,
			})
			if err != nil {
				if cerr, ok := err.(*conflict.ConflictError); ok {
					cerr.Source = f.Name
					cerr.OpType = "unzip"
					return cerr
				}
				return err
			}

			if entryResolvedPath == "" {
				processedFiles++
				continue // Skip
			}
			if entryResolvedPath == fpath && opts.Conflict.Policy == conflict.Overwrite {
				if err := opts.OpCtx.FS.RemoveAll(opts.OpCtx.Context, fpath); err != nil {
					logger.Warnf("Failed to remove existing file for unzip overwrite: %v", err)
				}
			}

			// Ensure parent directory exists
			parent := filepath.Dir(entryResolvedPath)
			if err := opts.OpCtx.FS.MkdirAll(opts.OpCtx.Context, parent, 0755); err != nil {
				return errors.WrapErrorWithPath(err, "MkdirAllParent", parent)
			}

			// Extract file
			rc, err := f.Open()
			if err != nil {
				return errors.WrapErrorWithPath(err, "OpenZipEntry", f.Name)
			}

			out, err := opts.OpCtx.FS.Create(opts.OpCtx.Context, entryResolvedPath)
			if err != nil {
				rc.Close()
				return errors.WrapErrorWithPath(err, "CreateExtract", entryResolvedPath)
			}

			_, err = io.Copy(out, rc)
			out.Close()
			rc.Close()

			if err != nil {
				return errors.WrapErrorWithPath(err, "CopyExtract", entryResolvedPath)
			}

			if err := opts.OpCtx.FS.Chmod(opts.OpCtx.Context, entryResolvedPath, f.Mode()); err != nil {
				logger.Warnf("Failed to set permissions on extracted file %s: %v", entryResolvedPath, err)
			}

			processedFiles++
			if opts.OpCtx.Progress != nil {
				label := fmt.Sprintf("Extracting %d/%d: %s", processedFiles, totalFiles, f.Name)
				if isRenamed {
					label = fmt.Sprintf("Extracting %d/%d: %s as %s", processedFiles, totalFiles, f.Name, opts.OpCtx.FS.Base(entryResolvedPath))
				}
				opts.OpCtx.Progress <- core.Progress{
					Label:   label,
					Percent: float64(processedFiles) / float64(totalFiles),
				}
			}
		}
	}

	return nil
}
