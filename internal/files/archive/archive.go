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

	"fm/internal/files/core"
	"fm/internal/files/errors"
)

// baseArchiveFS provides shared path resolution and metadata logic for archives.
type baseArchiveFS struct {
	archivePath string
}

func (b *baseArchiveFS) Separator() string {
	return "/"
}

func (b *baseArchiveFS) Join(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}

func (b *baseArchiveFS) Abs(path string) (string, error) {
	if filepath.IsAbs(path) {
		return b.Clean(path), nil
	}
	return b.Join("/", path), nil
}

func (b *baseArchiveFS) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
}

func (b *baseArchiveFS) Clean(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func (b *baseArchiveFS) Dir(path string) string {
	return filepath.ToSlash(filepath.Dir(path))
}

func (b *baseArchiveFS) Base(path string) string {
	return filepath.Base(path)
}

func (b *baseArchiveFS) Ext(path string) string {
	return filepath.Ext(path)
}

func (b *baseArchiveFS) GetHomeDir() (string, error) {
	return "/", nil
}

func (b *baseArchiveFS) IsLocal() bool {
	return false
}

func (b *baseArchiveFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	return true, nil
}

func (b *baseArchiveFS) Address() string {
	return b.archivePath
}

func (b *baseArchiveFS) User() string {
	return ""
}

// ZipFS implements FileSystem for a zip archive.
type ZipFS struct {
	baseArchiveFS
	reader *zip.ReadCloser
}

// NewArchiveFS creates a new FileSystem from an archive file path.
func NewArchiveFS(path string) (core.FileSystem, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".zip" {
		return NewZipFS(path)
	}
	if ext == ".tar" || ext == ".gz" || ext == ".tgz" {
		return NewTarFS(path)
	}
	return nil, fmt.Errorf("unsupported archive format: %s", ext)
}

// NewZipFS creates a new ZipFS from a zip file path.
func NewZipFS(path string) (*ZipFS, error) {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return nil, errors.WrapErrorWithPath(err, "OpenArchive", path)
	}

	return &ZipFS{
		baseArchiveFS: baseArchiveFS{archivePath: path},
		reader:        rc,
	}, nil
}

func (a *ZipFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
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

func (a *ZipFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
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

func (a *ZipFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
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

func (a *ZipFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return a.Stat(ctx, path)
}

func (a *ZipFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
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

func (a *ZipFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	return nil, fmt.Errorf("archive filesystem is read-only")
}

func (a *ZipFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *ZipFS) RemoveAll(ctx context.Context, path string) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *ZipFS) Rename(ctx context.Context, oldPath, newPath string) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *ZipFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *ZipFS) Preallocate(ctx context.Context, path string, size int64) error {
	return fmt.Errorf("archive filesystem is read-only")
}

func (a *ZipFS) Walk(ctx context.Context, root string, walkFn func(path string, info os.FileInfo, err error) error) error {
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

func (a *ZipFS) Close() error {
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
