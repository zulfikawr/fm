package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/logger"
)

// TarFS implements FileSystem for a tar archive.
type TarFS struct {
	baseArchiveFS
	entries []tarEntry
}

type tarEntry struct {
	header *tar.Header
}

// NewTarFS creates a new TarFS from a tar file path.
func NewTarFS(path string) (*TarFS, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "OpenArchive", path)
	}
	defer logger.CloseAndLog(f, "tar archive file for initialization")

	var tr *tar.Reader
	if strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".tgz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			return nil, errors.WrapErrorWithPath(err, "OpenArchive", path)
		}
		defer logger.CloseAndLog(gzr, "gzip reader for tar archive")
		tr = tar.NewReader(gzr)
	} else {
		tr = tar.NewReader(f)
	}

	var entries []tarEntry
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.WrapErrorWithPath(err, "ReadArchive", path)
		}
		entries = append(entries, tarEntry{
			header: header,
		})
	}

	return &TarFS{
		baseArchiveFS: baseArchiveFS{archivePath: path},
		entries:       entries,
	}, nil
}

func (fs *TarFS) Close() error {
	return nil
}

func (fs *TarFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	entries, err := fs.ReadDirEntries(ctx, path)
	if err != nil {
		return nil, err
	}

	infos := make([]os.FileInfo, len(entries))
	for i, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, errors.WrapErrorWithPath(err, "ReadDir.Info", path)
		}
		infos[i] = info
	}
	return infos, nil
}

func (fs *TarFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path = fs.Clean(path)
	if path == "." || path == "/" {
		path = ""
	} else {
		path = strings.TrimPrefix(path, "/")
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
	}

	seen := make(map[string]bool)
	var entries []os.DirEntry

	for i := range fs.entries {
		entry := fs.entries[i]
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		name := entry.header.Name
		if path != "" && !strings.HasPrefix(name, path) {
			continue
		}

		rel := strings.TrimPrefix(name, path)
		if rel == "" {
			continue
		}

		parts := strings.Split(rel, "/")
		base := parts[0]

		if seen[base] {
			continue
		}
		seen[base] = true

		isDir := len(parts) > 1 || entry.header.Typeflag == tar.TypeDir
		entries = append(entries, &tarDirEntry{
			name:  base,
			isDir: isDir,
			entry: entry,
		})
	}

	return entries, nil
}

func (fs *TarFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path = fs.Clean(path)
	if path == "." || path == "/" || path == "" {
		return &archiveFileInfo{
			name:  fs.Base(fs.archivePath),
			isDir: true,
		}, nil
	}

	path = strings.TrimPrefix(path, "/")

	for i := range fs.entries {
		entry := fs.entries[i]
		if strings.TrimSuffix(entry.header.Name, "/") == path {
			return entry.header.FileInfo(), nil
		}
	}

	// Check if it's a directory
	dirPath := path
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}
	for i := range fs.entries {
		entry := fs.entries[i]
		if strings.HasPrefix(entry.header.Name, dirPath) {
			return &archiveFileInfo{
				name:  fs.Base(path),
				isDir: true,
			}, nil
		}
	}

	return nil, errors.WrapErrorWithPath(os.ErrNotExist, "Stat", path)
}

func (fs *TarFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return fs.Stat(ctx, path)
}

func (fs *TarFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	f, err := os.Open(fs.archivePath)
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "OpenArchive", fs.archivePath)
	}

	var tr *tar.Reader
	var gzr *gzip.Reader
	if strings.HasSuffix(fs.archivePath, ".gz") || strings.HasSuffix(fs.archivePath, ".tgz") {
		var err error
		gzr, err = gzip.NewReader(f)
		if err != nil {
			logger.CloseAndLog(f, "tar archive file on gzr error")
			return nil, errors.WrapErrorWithPath(err, "OpenArchive", fs.archivePath)
		}
		tr = tar.NewReader(gzr)
	} else {
		tr = tar.NewReader(f)
	}

	path = strings.TrimPrefix(fs.Clean(path), "/")
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			if gzr != nil {
				logger.CloseAndLog(gzr, "gzip reader on read error")
			}
			logger.CloseAndLog(f, "tar archive file on read error")
			return nil, errors.WrapErrorWithPath(err, "ReadArchive", fs.archivePath)
		}

		if strings.TrimSuffix(header.Name, "/") == path {
			return &tarFileReadCloser{
				Reader: tr,
				f:      f,
				gzr:    gzr,
			}, nil
		}
	}

	if gzr != nil {
		logger.CloseAndLog(gzr, "gzip reader on path not found")
	}
	logger.CloseAndLog(f, "tar archive file on path not found")
	return nil, errors.WrapErrorWithPath(os.ErrNotExist, "Open", path)
}

func (fs *TarFS) Walk(ctx context.Context, root string, walkFn filepath.WalkFunc) error {
	root = strings.TrimPrefix(fs.Clean(root), "/")
	if root == "." {
		root = ""
	}

	for i := range fs.entries {
		entry := fs.entries[i]
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if root != "" && !strings.HasPrefix(entry.header.Name, root) {
			continue
		}
		if err := walkFn(entry.header.Name, entry.header.FileInfo(), nil); err != nil {
			return err
		}
	}
	return nil
}

// Writer operations (Read-Only)

func (fs *TarFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	return nil, &errors.UnsupportedOperationError{Op: "Create", Filesystem: "Tar"}
}

func (fs *TarFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	return &errors.UnsupportedOperationError{Op: "MkdirAll", Filesystem: "Tar"}
}

func (fs *TarFS) RemoveAll(ctx context.Context, path string) error {
	return &errors.UnsupportedOperationError{Op: "RemoveAll", Filesystem: "Tar"}
}

func (fs *TarFS) Rename(ctx context.Context, oldPath, newPath string) error {
	return &errors.UnsupportedOperationError{Op: "Rename", Filesystem: "Tar"}
}

func (fs *TarFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return &errors.UnsupportedOperationError{Op: "Chmod", Filesystem: "Tar"}
}

func (fs *TarFS) Preallocate(ctx context.Context, path string, size int64) error {
	return &errors.UnsupportedOperationError{Op: "Preallocate", Filesystem: "Tar"}
}

// Helper structs

type tarDirEntry struct {
	name  string
	isDir bool
	entry tarEntry
}

func (e *tarDirEntry) Name() string               { return e.name }
func (e *tarDirEntry) IsDir() bool                { return e.isDir }
func (e *tarDirEntry) Type() os.FileMode          { return e.entry.header.FileInfo().Mode().Type() }
func (e *tarDirEntry) Info() (os.FileInfo, error) { return e.entry.header.FileInfo(), nil }

type tarFileReadCloser struct {
	io.Reader
	f   *os.File
	gzr *gzip.Reader
}

func (rc *tarFileReadCloser) Close() error {
	if rc.gzr != nil {
		logger.LogIfError(rc.gzr.Close(), "failed to close gzip reader in tarFileReadCloser")
	}
	return rc.f.Close()
}
