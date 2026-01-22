package core

// IsRoot returns true if the given path is the root of the filesystem.
func IsRoot(fs FileSystem, path string) bool {
	if path == "" {
		return true
	}
	return fs.Dir(path) == path
}

// GetParent returns the parent directory of the given path.
func GetParent(fs FileSystem, path string) string {
	return fs.Dir(path)
}
