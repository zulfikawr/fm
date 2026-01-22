package core

import (
	"path/filepath"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestFileSystemHelpers(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	fs.DirFunc = func(path string) string {
		if path == "/" {
			return "/"
		}
		return filepath.Dir(path)
	}

	t.Run("IsRoot", func(t *testing.T) {
		if !IsRoot(fs, "/") {
			t.Error("IsRoot(/) should be true")
		}
		if IsRoot(fs, "/some/path") {
			t.Error("IsRoot(/some/path) should be false")
		}
	})

	t.Run("GetParent", func(t *testing.T) {
		parent := GetParent(fs, "/some/path")
		if parent != "/some" {
			t.Errorf("GetParent failed, expected /some, got %s", parent)
		}
	})
}
