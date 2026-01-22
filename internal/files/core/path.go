package core

import (
	"path"
	"path/filepath"
	"strings"
)

// NativePathResolver uses the OS-specific filepath package.
type NativePathResolver struct{}

func (n NativePathResolver) Separator() string                     { return string(filepath.Separator) }
func (n NativePathResolver) Join(elem ...string) string            { return filepath.Join(elem...) }
func (n NativePathResolver) Abs(p string) (string, error)          { return filepath.Abs(p) }
func (n NativePathResolver) Rel(base, targ string) (string, error) { return filepath.Rel(base, targ) }
func (n NativePathResolver) Clean(p string) string                 { return filepath.Clean(p) }
func (n NativePathResolver) Dir(p string) string                   { return filepath.Dir(p) }
func (n NativePathResolver) Base(p string) string                  { filepath.Base(p); return filepath.Base(p) }
func (n NativePathResolver) Ext(p string) string                   { return filepath.Ext(p) }

// UnixPathResolver uses the Unix-style path package.
type UnixPathResolver struct{}

func (u UnixPathResolver) Separator() string          { return "/" }
func (u UnixPathResolver) Join(elem ...string) string { return path.Join(elem...) }
func (u UnixPathResolver) Abs(p string) (string, error) {
	if path.IsAbs(p) {
		return path.Clean(p), nil
	}
	return p, nil // Abs for UnixPathResolver usually needs context, implementation should override if needed.
}
func (u UnixPathResolver) Rel(base, targ string) (string, error) {
	return RelUnix(base, targ)
}
func (u UnixPathResolver) Clean(p string) string { return path.Clean(p) }
func (u UnixPathResolver) Dir(p string) string   { return path.Dir(p) }
func (u UnixPathResolver) Base(p string) string  { return path.Base(p) }
func (u UnixPathResolver) Ext(p string) string   { return path.Ext(p) }

// RelUnix is a helper for Unix-style path relativity as the 'path' package lacks it.
func RelUnix(basepath, targpath string) (string, error) {
	base := path.Clean(basepath)
	targ := path.Clean(targpath)

	if base == targ {
		return ".", nil
	}

	baseElems := strings.Split(strings.Trim(base, "/"), "/")
	targElems := strings.Split(strings.Trim(targ, "/"), "/")

	if base == "/" {
		baseElems = []string{}
	}
	if targ == "/" {
		targElems = []string{}
	}

	i := 0
	for i < len(baseElems) && i < len(targElems) && baseElems[i] == targElems[i] {
		i++
	}

	var rel []string
	for j := i; j < len(baseElems); j++ {
		rel = append(rel, "..")
	}
	rel = append(rel, targElems[i:]...)

	if len(rel) == 0 {
		return ".", nil
	}

	return strings.Join(rel, "/"), nil
}
