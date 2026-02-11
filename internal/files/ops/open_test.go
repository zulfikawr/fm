package ops

import (
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
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
		result := IsTextFile(fs, tt.path)
		if result != tt.expected {
			t.Errorf("IsTextFile(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestGetOpenCmd(t *testing.T) {
	fs := testutil.NewMockFileSystem()

	// Mock lookPath to always succeed
	oldLookPath := lookPath
	lookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}
	defer func() { lookPath = oldLookPath }()

	t.Run("Terminal Editor", func(t *testing.T) {
		// Index 0 is vim based on constants.Editors
		cmd, isTerm, err := GetOpenCmd(OpenOptions{
			FS:        fs,
			Path:      "file.txt",
			EditorIdx: 0,
		})
		testutil.AssertNoError(t, err, "Should get cmd")
		testutil.AssertEqual(t, true, isTerm, "Vim should be terminal editor")
		if !strings.HasSuffix(cmd.Path, "vim") {
			t.Errorf("Cmd path %q should end with vim", cmd.Path)
		}
	})

	t.Run("Open at line", func(t *testing.T) {
		// Vim/Nano/Vi
		cmd, isTerm, err := GetOpenAtLineCmd(OpenOptions{
			FS:        fs,
			Path:      "file.txt",
			EditorIdx: 0,
			Line:      10,
		})
		testutil.AssertNoError(t, err, "Should get vim cmd")
		testutil.AssertEqual(t, true, isTerm, "Vim should be terminal")
		testutil.AssertEqual(t, "+10", cmd.Args[1], "Vim should use +10")

		// VS Code (at index 5 based on constants.Editors)
		cmd, isTerm, err = GetOpenAtLineCmd(OpenOptions{
			FS:        fs,
			Path:      "file.txt",
			EditorIdx: 5,
			Line:      10,
		})
		testutil.AssertNoError(t, err, "Should get code cmd")
		testutil.AssertEqual(t, false, isTerm, "Code should not be terminal")
		testutil.AssertEqual(t, "--goto", cmd.Args[1], "VS Code should use --goto")
		testutil.AssertEqual(t, "file.txt:10", cmd.Args[2], "VS Code should use path:line")

		// Default (no line)
		cmd, isTerm, err = GetOpenAtLineCmd(OpenOptions{
			FS:        fs,
			Path:      "file.txt",
			EditorIdx: 0,
			Line:      0,
		})
		testutil.AssertNoError(t, err, "Should get default cmd")
		testutil.AssertEqual(t, true, isTerm, "Vim should be terminal")
		testutil.AssertEqual(t, 2, len(cmd.Args), "Should have 2 args (cmd + path)")
	})

	t.Run("Non-text file", func(t *testing.T) {
		_, isTerm, err := GetOpenCmd(OpenOptions{
			FS:        fs,
			Path:      "image.png",
			EditorIdx: 0,
		})
		testutil.AssertNoError(t, err, "Should get open cmd for image")
		testutil.AssertEqual(t, true, isTerm, "Image should be opened in terminal editor if vim selected")
	})
}

func TestIsTerminalEditor(t *testing.T) {
	testutil.AssertEqual(t, true, isTerminalEditor("vim"), "vim is terminal editor")
	testutil.AssertEqual(t, true, isTerminalEditor("nano"), "nano is terminal editor")
	testutil.AssertEqual(t, false, isTerminalEditor("code"), "code is NOT terminal editor")
}
