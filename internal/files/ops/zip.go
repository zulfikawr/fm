package ops

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fm/internal/files/conflict"
	"fm/internal/files/core"
	"fm/internal/files/errors"
	"fm/internal/logger"
)

// Zip compresses multiple files or directories into a single zip archive.
func Zip(ctx context.Context, fs core.FileSystem, sources []string, dst string, progChan chan<- core.Progress, policy conflict.Policy) error {
	if len(sources) == 0 {
		return errors.WrapError(fmt.Errorf("no source files specified"), "Zip")
	}

	resolver := conflict.NewResolver()
	resolvedDst, _, err := resolver.Resolve(ctx, fs, sources[0], dst, policy)
	if err != nil {
		if cerr, ok := err.(*conflict.ConflictError); ok {
			cerr.PendingItems = sources
			cerr.OpType = "zip"
			return cerr
		}
		return err
	}

	if resolvedDst == "" {
		return nil // Skip
	}
	dst = resolvedDst

	// Create destination file
	out, err := fs.Create(ctx, dst)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Create", dst)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	// For simplicity, we'll increment progress per file processed
	processedFiles := 0

	for _, src := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		baseDir := fs.Dir(src)

		err = walkAndZip(ctx, fs, src, baseDir, zw, func() {
			processedFiles++
			if progChan != nil {
				progChan <- core.Progress{
					Label:   fmt.Sprintf("Zipping %d items into %s", len(sources), fs.Base(dst)),
					Percent: float64(processedFiles) / float64(len(sources)), // Simplified progress
				}
			}
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func walkAndZip(ctx context.Context, fs core.FileSystem, currentPath, baseDir string, zw *zip.Writer, onFile func()) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	info, err := fs.Lstat(ctx, currentPath)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", currentPath)
	}

	// Calculate relative path for the zip header
	relPath, err := fs.Rel(baseDir, currentPath)
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
		_, err = zw.Create(relPath)
		if err != nil {
			return errors.WrapErrorWithPath(err, "ZipCreateDir", relPath)
		}

		entries, err := fs.ReadDir(ctx, currentPath)
		if err != nil {
			return errors.WrapErrorWithPath(err, "ReadDir", currentPath)
		}

		for _, entry := range entries {
			childPath := fs.Join(currentPath, entry.Name())
			if err := walkAndZip(ctx, fs, childPath, baseDir, zw, onFile); err != nil {
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

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return errors.WrapErrorWithPath(err, "ZipCreate", currentPath)
	}

	in, err := fs.Open(ctx, currentPath)
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

	onFile()
	return nil
}

// Unzip extracts a zip archive to the specified destination directory.
func Unzip(ctx context.Context, fs core.FileSystem, src, dst string, progChan chan<- core.Progress, policy conflict.Policy) error {
	resolver := conflict.NewResolver()
	resolvedDst, _, err := resolver.Resolve(ctx, fs, src, dst, policy)
	if err != nil {
		if cerr, ok := err.(*conflict.ConflictError); ok {
			cerr.PendingItems = []string{src}
			cerr.OpType = "unzip"
			return cerr
		}
		return err
	}

	if resolvedDst == "" {
		return nil // Skip
	}
	dst = resolvedDst

	// Open the source file
	in, err := fs.Open(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Open", src)
	}
	defer in.Close()

	// zip.NewReader needs a ReaderAt and the size.
	// We'll get the size first.
	info, err := fs.Stat(ctx, src)
	if err != nil {
		return errors.WrapErrorWithPath(err, "Stat", src)
	}
	size := info.Size()

	// If it's a local file, we can use os.Open directly for ReaderAt
	var readerAt io.ReaderAt
	if fs.IsLocal() {
		f, err := os.Open(src)
		if err != nil {
			return errors.WrapErrorWithPath(err, "OpenLocal", src)
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

	var isRenamed bool

	for _, f := range zr.File {

		select {

		case <-ctx.Done():

			return ctx.Err()

		default:

		}

		fpath, err := conflict.ValidateSecurePath(fs, dst, f.Name)

		if err != nil {

			continue // Skip dangerous paths

		}

		if f.FileInfo().IsDir() {

			err = fs.MkdirAll(ctx, fpath, f.Mode()|0111) // Ensure execute permission for directory access

			if err != nil {

				return errors.WrapErrorWithPath(err, "MkdirAll", fpath)

			}

		} else {

			// Check for conflict

			var resolvedPath string

			resolvedPath, isRenamed, err = resolver.Resolve(ctx, fs, f.Name, fpath, policy)

			if err != nil {

				if cerr, ok := err.(*conflict.ConflictError); ok {
					// For Unzip, we might want to return the error to the UI
					// but it's tricky because we are inside a zip archive.
					// We'll mark the source as the entry name.
					cerr.Source = f.Name
					cerr.OpType = "unzip"
					// We can't easily provide remaining items here without complex state
					return cerr
				}
				return err
			}

			if resolvedPath == "" {
				processedFiles++
				continue // Skip
			}
			if resolvedPath == fpath && policy == conflict.Overwrite {
				if err := fs.RemoveAll(ctx, fpath); err != nil {
					logger.Warnf("Failed to remove existing file for unzip overwrite: %v", err)
				}
			}
			fpath = resolvedPath

			// Ensure parent directory exists
			parent := filepath.Dir(fpath)
			if err := fs.MkdirAll(ctx, parent, 0755); err != nil {
				return errors.WrapErrorWithPath(err, "MkdirAllParent", parent)
			}

			// Extract file
			rc, err := f.Open()
			if err != nil {
				return errors.WrapErrorWithPath(err, "OpenZipEntry", f.Name)
			}

			out, err := fs.Create(ctx, fpath)
			if err != nil {
				rc.Close()
				return errors.WrapErrorWithPath(err, "CreateExtract", fpath)
			}

			_, err = io.Copy(out, rc)
			out.Close()
			rc.Close()

			if err != nil {
				return errors.WrapErrorWithPath(err, "CopyExtract", fpath)
			}

			if err := fs.Chmod(ctx, fpath, f.Mode()); err != nil {
				logger.Warnf("Failed to set permissions on extracted file %s: %v", fpath, err)
			}
		}

		processedFiles++
		if progChan != nil {
			label := fmt.Sprintf("Extracting %d/%d: %s", processedFiles, totalFiles, f.Name)
			if isRenamed {
				label = fmt.Sprintf("Extracting %d/%d: %s as %s", processedFiles, totalFiles, f.Name, fs.Base(fpath))
			}
			progChan <- core.Progress{
				Label:   label,
				Percent: float64(processedFiles) / float64(totalFiles),
			}
		}
	}

	return nil
}
