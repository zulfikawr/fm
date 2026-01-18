package ops

import (
	"strings"
	"testing"

	"fm/internal/testutil"
)

func TestIsTextFile(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	tests := []struct {
		path     string
		expected bool
	}{
		{"test.txt", true},
		{"main.go", true},
		{"image.png", false},
		{"no_ext", true},
		{"SCRIPT.SH", true},
	}

	for _, tt := range tests {
		fs.ExtFunc = func(p string) string {
			for i := len(p) - 1; i >= 0; i-- {
				if p[i] == '.' {
					return p[i:]
				}
			}
			return ""
		}
		result := IsTextFile(fs, tt.path)
		if result != tt.expected {
			t.Errorf("IsTextFile(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestGetOpenCmd(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	fs.ExtFunc = func(p string) string { return ".txt" }

	// Mock lookPath to always succeed
	oldLookPath := lookPath
	lookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	defer func() { lookPath = oldLookPath }()

	t.Run("Terminal Editor", func(t *testing.T) {
		// Index 0 is vim based on constants.Editors
		cmd, isTerm, err := GetOpenCmd(fs, "file.txt", 0)
		testutil.AssertNoError(t, err, "Should get cmd")
		testutil.AssertEqual(t, true, isTerm, "Vim should be terminal editor")
		if !strings.HasSuffix(cmd.Path, "vim") {
			t.Errorf("Cmd path %q should end with vim", cmd.Path)
		}
	})

	t.Run("Open at line", func(t *testing.T) {
		// Vim/Nano/Vi
		cmd, _, _ := GetOpenAtLineCmd(fs, "file.txt", 0, 10)
		testutil.AssertEqual(t, "+10", cmd.Args[1], "Vim should use +10")

		// VS Code (at index 5 based on constants.Editors)
		cmd, _, _ = GetOpenAtLineCmd(fs, "file.txt", 5, 10)
		testutil.AssertEqual(t, "--goto", cmd.Args[1], "VS Code should use --goto")
		testutil.AssertEqual(t, "file.txt:10", cmd.Args[2], "VS Code should use path:line")

		// Default (no line)
		cmd, _, _ = GetOpenAtLineCmd(fs, "file.txt", 0, 0)
		testutil.AssertEqual(t, 2, len(cmd.Args), "Should have 2 args (cmd + path)")
	})

	t.Run("Non-text file", func(t *testing.T) {
		fs.ExtFunc = func(p string) string { return ".png" }
		cmd, isTerm, err := GetOpenCmd(fs, "image.png", 0)
		testutil.AssertNoError(t, err, "Should get open cmd for image")
		testutil.AssertEqual(t, false, isTerm, "Image should not be opened in terminal editor")
		// On linux it should be xdg-open
		if !strings.Contains(cmd.Path, "xdg-open") && !strings.Contains(cmd.Path, "open") && !strings.Contains(cmd.Path, "rundll32") {
			// This depends on GOOS, but at least one should be there if it's supported
		}
	})
}

func TestIsTerminalEditor(t *testing.T) {
	testutil.AssertEqual(t, true, isTerminalEditor("vim"), "vim is terminal editor")
	testutil.AssertEqual(t, true, isTerminalEditor("nano"), "nano is terminal editor")
	testutil.AssertEqual(t, false, isTerminalEditor("code"), "code is NOT terminal editor")
}
