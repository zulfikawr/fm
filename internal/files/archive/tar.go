package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"fm/internal/files/errors"
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
	defer f.Close()

	var tr *tar.Reader
	if strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".tgz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			return nil, errors.WrapErrorWithPath(err, "OpenArchive", path)
		}
		defer gzr.Close()
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

func (a *TarFS) Close() error {
	return nil
}

func (a *TarFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	entries, err := a.ReadDirEntries(ctx, path)
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

func (a *TarFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path = a.Clean(path)
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

	for _, entry := range a.entries {
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

func (a *TarFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path = a.Clean(path)
	if path == "." || path == "/" || path == "" {
		return &archiveFileInfo{
			name:  a.Base(a.archivePath),
			isDir: true,
		}, nil
	}

	path = strings.TrimPrefix(path, "/")

	for _, entry := range a.entries {
		if strings.TrimSuffix(entry.header.Name, "/") == path {
			return entry.header.FileInfo(), nil
		}
	}

	// Check if it's a directory
	dirPath := path
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}
	for _, entry := range a.entries {
		if strings.HasPrefix(entry.header.Name, dirPath) {
			return &archiveFileInfo{
				name:  a.Base(path),
				isDir: true,
			}, nil
		}
	}

	return nil, errors.WrapErrorWithPath(os.ErrNotExist, "Stat", path)
}

func (a *TarFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return a.Stat(ctx, path)
}

func (a *TarFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	f, err := os.Open(a.archivePath)
	if err != nil {
		return nil, err
	}

	var tr *tar.Reader
	if strings.HasSuffix(a.archivePath, ".gz") || strings.HasSuffix(a.archivePath, ".tgz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		tr = tar.NewReader(gzr)
	} else {
		tr = tar.NewReader(f)
	}

	path = strings.TrimPrefix(a.Clean(path), "/")
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			f.Close()
			return nil, err
		}

		if strings.TrimSuffix(header.Name, "/") == path {
			return &tarFileReadCloser{
				Reader: tr,
				f:      f,
			}, nil
		}
	}

	f.Close()
	return nil, errors.WrapErrorWithPath(os.ErrNotExist, "Open", path)
}

func (a *TarFS) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	root = strings.TrimPrefix(a.Clean(root), "/")
	if root == "." {
		root = ""
	}

	for _, entry := range a.entries {
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

func (a *TarFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	return nil, fmt.Errorf("archive filesystem is read-only")
}

func (a *TarFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *TarFS) RemoveAll(ctx context.Context, path string) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *TarFS) Rename(ctx context.Context, oldPath, newPath string) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *TarFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *TarFS) Preallocate(ctx context.Context, path string, size int64) error {
	return fmt.Errorf("archive filesystem is read-only")
}

// Helper structs

type tarDirEntry struct {
	name  string
	isDir bool
	entry tarEntry
}

func (d *tarDirEntry) Name() string               { return d.name }
func (d *tarDirEntry) IsDir() bool                { return d.isDir }
func (d *tarDirEntry) Type() os.FileMode          { return d.entry.header.FileInfo().Mode().Type() }
func (d *tarDirEntry) Info() (os.FileInfo, error) { return d.entry.header.FileInfo(), nil }

type tarFileReadCloser struct {
	io.Reader
	f *os.File
}

func (t *tarFileReadCloser) Close() error {
	return t.f.Close()
}
