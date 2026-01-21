package core

import (
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestPathHelpers(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	fs.DirFunc = func(path string) string {
		if path == "/" {
			return "/"
		}
		if path == "/a" {
			return "/"
		}
		if path == "C:\\" {
			return "C:\\"
		}
		return "/"
	}

	t.Run("IsRoot", func(t *testing.T) {
		testCases := []struct {
			path     string
			expected bool
		}{
			{"/", true},
			{"/a", false},
			{"", true},
			{"C:\\", true},
		}

		for _, tc := range testCases {
			testutil.AssertEqual(t, tc.expected, IsRoot(fs, tc.path), tc.path)
		}
	})

	t.Run("GetParent", func(t *testing.T) {
		testutil.AssertEqual(t, "/", GetParent(fs, "/a"), "Parent of /a")
	})
}
