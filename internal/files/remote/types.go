// Package remote implements the FileSystem interface for SFTP/SSH remote filesystems.
// It provides connection management, automatic reconnection, and parallel directory traversal.
package remote

import (
	"os"
)

type dirEntry struct {
	info os.FileInfo
}

func (d *dirEntry) Name() string               { return d.info.Name() }
func (d *dirEntry) IsDir() bool                { return d.info.IsDir() }
func (d *dirEntry) Type() os.FileMode          { return d.info.Mode().Type() }
func (d *dirEntry) Info() (os.FileInfo, error) { return d.info, nil }

func infoToDirEntry(info os.FileInfo) os.DirEntry {
	return &dirEntry{info: info}
}
