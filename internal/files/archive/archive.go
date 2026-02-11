package archive

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
)

// baseArchiveFS provides shared path resolution and metadata logic for archives.
type baseArchiveFS struct {
	archivePath string
}

func (fs *baseArchiveFS) Separator() string {
	return "/"
}

func (fs *baseArchiveFS) Join(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}

func (fs *baseArchiveFS) Abs(path string) (string, error) {
	if filepath.IsAbs(path) {
		return fs.Clean(path), nil
	}
	return fs.Join("/", path), nil
}

func (fs *baseArchiveFS) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
}

func (fs *baseArchiveFS) Clean(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func (fs *baseArchiveFS) Dir(path string) string {
	return filepath.ToSlash(filepath.Dir(path))
}

func (fs *baseArchiveFS) Base(path string) string {
	return filepath.Base(path)
}

func (fs *baseArchiveFS) Ext(path string) string {
	return filepath.Ext(path)
}

func (fs *baseArchiveFS) GetHomeDir() (string, error) {
	return "/", nil
}

func (fs *baseArchiveFS) IsLocal() bool {
	return false
}

func (fs *baseArchiveFS) IsReadOnly(ctx context.Context, path string) (bool, error) {
	return true, nil
}

func (fs *baseArchiveFS) Address() string {
	return fs.archivePath
}

func (fs *baseArchiveFS) User() string {
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
	return nil, &errors.ValidationError{
		Field:   "extension",
		Value:   ext,
		Message: "unsupported archive format",
	}
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

func (fs *ZipFS) ReadDir(ctx context.Context, path string) ([]os.FileInfo, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

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

func (fs *ZipFS) ReadDirEntries(ctx context.Context, path string) ([]os.DirEntry, error) {
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

	files := fs.reader.File
	for i := range files {
		file := files[i]
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

func (fs *ZipFS) Stat(ctx context.Context, path string) (os.FileInfo, error) {
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

	// Check for exact file match
	files := fs.reader.File
	for i := range files {
		file := files[i]
		if strings.TrimSuffix(file.Name, "/") == path {
			return file.FileInfo(), nil
		}
	}

	// Check if it's a directory
	dirPath := path
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}
	for i := range files {
		file := files[i]
		if strings.HasPrefix(file.Name, dirPath) {
			return &archiveFileInfo{
				name:  fs.Base(path),
				isDir: true,
			}, nil
		}
	}

	return nil, errors.WrapErrorWithPath(os.ErrNotExist, "Stat", path)
}

func (fs *ZipFS) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return fs.Stat(ctx, path)
}

func (fs *ZipFS) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	path = strings.TrimPrefix(fs.Clean(path), "/")
	files := fs.reader.File
	for i := range files {
		file := files[i]
		if file.Name == path {
			f, err := file.Open()
			return f, errors.WrapErrorWithPath(err, "Open", path)
		}
	}
	return nil, errors.WrapErrorWithPath(os.ErrNotExist, "Open", path)
}

// Writer operations (Read-Only)

func (fs *ZipFS) Create(ctx context.Context, path string) (io.WriteCloser, error) {
	return nil, &errors.UnsupportedOperationError{Op: "Create", Filesystem: "Zip"}
}

func (fs *ZipFS) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	return &errors.UnsupportedOperationError{Op: "MkdirAll", Filesystem: "Zip"}
}

func (fs *ZipFS) RemoveAll(ctx context.Context, path string) error {
	return &errors.UnsupportedOperationError{Op: "RemoveAll", Filesystem: "Zip"}
}

func (fs *ZipFS) Rename(ctx context.Context, oldPath, newPath string) error {
	return &errors.UnsupportedOperationError{Op: "Rename", Filesystem: "Zip"}
}

func (fs *ZipFS) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return &errors.UnsupportedOperationError{Op: "Chmod", Filesystem: "Zip"}
}

func (fs *ZipFS) Preallocate(ctx context.Context, path string, size int64) error {
	return &errors.UnsupportedOperationError{Op: "Preallocate", Filesystem: "Zip"}
}

func (fs *ZipFS) Walk(ctx context.Context, root string, walkFn filepath.WalkFunc) error {
	// 1. Resolve root entry
	info, err := fs.Stat(ctx, root)
	if err != nil {
		return err
	}
	if info == nil {
		// Handle nil info
	}

	searchRoot := root
	if searchRoot == "/" {
		searchRoot = ""
	}
	searchRoot = strings.TrimPrefix(searchRoot, "/")

	files := fs.reader.File
	for i := range files {
		file := files[i]
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if searchRoot != "" && !strings.HasPrefix(file.Name, searchRoot) {
			continue
		}
		if err := walkFn(file.Name, file.FileInfo(), nil); err != nil {
			return err
		}
	}
	return nil
}

func (fs *ZipFS) Close() error {
	if fs.reader != nil {
		err := fs.reader.Close()
		return errors.WrapErrorWithPath(err, "CloseArchive", fs.archivePath)
	}
	return nil
}

// GetDefaultExtractionPath calculates a default folder name for extraction by stripping known archive extensions.
func GetDefaultExtractionPath(fs core.FileSystem, archivePath string) string {
	baseName := fs.Base(archivePath)
	ext := fs.Ext(baseName)
	name := strings.TrimSuffix(baseName, ext)

	// Handle double extensions like .tar.gz
	if before, ok := strings.CutSuffix(name, ".tar"); ok {
		name = before
	}

	return fs.Join(fs.Dir(archivePath), name)
}

// Helper structs
type archiveDirEntry struct {
	name  string
	isDir bool
	file  *zip.File
}

func (e *archiveDirEntry) Name() string               { return e.name }
func (e *archiveDirEntry) IsDir() bool                { return e.isDir }
func (e *archiveDirEntry) Type() os.FileMode          { return e.file.Mode().Type() }
func (e *archiveDirEntry) Info() (os.FileInfo, error) { return e.file.FileInfo(), nil }

type archiveFileInfo struct {
	name  string
	isDir bool
}

func (info *archiveFileInfo) Name() string       { return info.name }
func (info *archiveFileInfo) Size() int64        { return 0 }
func (info *archiveFileInfo) Mode() os.FileMode  { return os.ModeDir | 0555 }
func (info *archiveFileInfo) ModTime() time.Time { return time.Time{} }
func (info *archiveFileInfo) IsDir() bool        { return info.isDir }
func (info *archiveFileInfo) Sys() interface{}   { return nil }
