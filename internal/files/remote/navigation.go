package remote

import (
	"path"
	"strings"

	"github.com/zulfikawr/fm/internal/files/errors"
)

func (fs *RemoteFS) Join(elem ...string) string {
	return path.Join(elem...)
}

func (fs *RemoteFS) Abs(p string) (string, error) {
	if path.IsAbs(p) {
		return path.Clean(p), nil
	}
	wd, err := fs.client.Getwd()
	if err != nil {
		return "", errors.WrapError(err, "Abs")
	}
	return path.Join(wd, p), nil
}

func (fs *RemoteFS) Rel(basepath, targpath string) (string, error) {
	base := path.Clean(basepath)
	targ := path.Clean(targpath)

	if base == targ {
		return ".", nil
	}

	// Simple implementation for slash-based paths
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

func (fs *RemoteFS) Clean(p string) string {
	return path.Clean(p)
}

func (fs *RemoteFS) Dir(p string) string {
	return path.Dir(p)
}

func (fs *RemoteFS) Base(p string) string {
	return path.Base(p)
}

func (fs *RemoteFS) Ext(p string) string {
	return path.Ext(p)
}
