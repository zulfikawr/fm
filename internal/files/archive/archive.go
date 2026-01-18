package archive

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fm/internal/files/errors"
)

// ArchiveFS implements FileSystem for a zip archive.
type ArchiveFS struct {
	archivePath string
	reader      *zip.ReadCloser
}

// NewArchiveFS creates a new ArchiveFS from a zip file path.
func NewArchiveFS(path string) (*ArchiveFS, error) {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "OpenArchive", path)
	}

	return &ArchiveFS{
		archivePath: path,
		reader:      rc,
	}, nil
}

func (a *ArchiveFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

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

func (a *ArchiveFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
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

	for _, file := range a.reader.File {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		name := file.Name
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

		isDir := len(parts) > 1 || strings.HasSuffix(rel, "/")
		entries = append(entries, &archiveDirEntry{
			name:  base,
			isDir: isDir,
			file:  file,
		})
	}

	return entries, nil
}

func (a *ArchiveFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
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

	// Check for exact file match
	for _, file := range a.reader.File {
		if strings.TrimSuffix(file.Name, "/") == path {
			return file.FileInfo(), nil
		}
	}

	// Check if it's a directory
	dirPath := path
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}
	for _, file := range a.reader.File {
		if strings.HasPrefix(file.Name, dirPath) {
			return &archiveFileInfo{
				name:  a.Base(path),
				isDir: true,
			}, nil
		}
	}

	return nil, errors.WrapErrorWithPath(os.ErrNotExist, "Stat", path)
}

func (a *ArchiveFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return a.Stat(ctx, path)
}

func (a *ArchiveFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path = strings.TrimPrefix(a.Clean(path), "/")
	for _, file := range a.reader.File {
		if file.Name == path {
			f, err := file.Open()
			return f, errors.WrapErrorWithPath(err, "Open", path)
		}
	}
	return nil, errors.WrapErrorWithPath(os.ErrNotExist, "Open", path)
}

// Writer operations (Read-Only)

func (a *ArchiveFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	return nil, fmt.Errorf("archive filesystem is read-only")
}

func (a *ArchiveFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *ArchiveFS) RemoveAll(ctx context.Context, path string) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *ArchiveFS) Rename(ctx context.Context, oldPath, newPath string) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *ArchiveFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *ArchiveFS) Preallocate(ctx context.Context, path string, size int64) error {
	return fmt.Errorf("archive filesystem is read-only")
}

// PathResolver

func (a *ArchiveFS) Separator() string {
	return "/"
}

func (a *ArchiveFS) Join(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}

func (a *ArchiveFS) Abs(path string) (string, error) {
	if filepath.IsAbs(path) {
		return a.Clean(path), nil
	}
	return a.Join("/", path), nil
}

func (a *ArchiveFS) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
}

func (a *ArchiveFS) Clean(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func (a *ArchiveFS) Dir(path string) string {
	return filepath.ToSlash(filepath.Dir(path))
}

func (a *ArchiveFS) Base(path string) string {
	return filepath.Base(path)
}

func (a *ArchiveFS) Ext(path string) string {
	return filepath.Ext(path)
}

func (a *ArchiveFS) GetHomeDir() (string, error) {
	return "/", nil
}

func (a *ArchiveFS) IsLocal() bool {
	return false
}

func (a *ArchiveFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	return true, nil
}

func (a *ArchiveFS) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	// Simple implementation: iterate over all files in reader
	root = strings.TrimPrefix(a.Clean(root), "/")
	if root == "." {
		root = ""
	}

	for _, file := range a.reader.File {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if root != "" && !strings.HasPrefix(file.Name, root) {
			continue
		}
		if err := walkFn(file.Name, file.FileInfo(), nil); err != nil {
			return err
		}
	}
	return nil
}

func (a *ArchiveFS) Address() string {
	return a.archivePath
}

func (a *ArchiveFS) User() string {
	return ""
}

func (a *ArchiveFS) Close() error {
	if a.reader != nil {
		err := a.reader.Close()
		return errors.WrapErrorWithPath(err, "CloseArchive", a.archivePath)
	}
	return nil
}

// Helper structs

type archiveDirEntry struct {
	name  string
	isDir bool
	file  *zip.File
}

func (d *archiveDirEntry) Name() string               { return d.name }
func (d *archiveDirEntry) IsDir() bool                { return d.isDir }
func (d *archiveDirEntry) Type() os.FileMode          { return d.file.Mode().Type() }
func (d *archiveDirEntry) Info() (os.FileInfo, error) { return d.file.FileInfo(), nil }

type archiveFileInfo struct {
	name  string
	isDir bool
}

func (f *archiveFileInfo) Name() string       { return f.name }
func (f *archiveFileInfo) Size() int64        { return 0 }
func (f *archiveFileInfo) Mode() os.FileMode  { return os.ModeDir | 0555 }
func (f *archiveFileInfo) ModTime() time.Time { return time.Time{} }
func (f *archiveFileInfo) IsDir() bool        { return f.isDir }
func (f *archiveFileInfo) Sys() interface{}   { return nil }
