package ops

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
)

var lookPath = exec.LookPath

// GetOpenCmd returns the command to open a file and a boolean indicating if it's a terminal-based editor.
func GetOpenCmd(fs core.FileSystem, path string, editorIdx int) (*exec.Cmd, bool, error) {
	return GetOpenAtLineCmd(fs, path, editorIdx, 0)
}

// GetOpenAtLineCmd returns a command to open a file at a specific line number.
func GetOpenAtLineCmd(fs core.FileSystem, path string, editorIdx int, line int) (*exec.Cmd, bool, error) {
	if IsTextFile(fs, path) {
		editor := constants.Editors[editorIdx]
		isTerminalEditor := isTerminalEditor(editor)

		// Check if editor exists
		if _, err := lookPath(editor); err != nil {
			return nil, false, errors.WrapError(err, "OpenWithEditor")
		}

		var args []string
		if line > 0 {
			switch editor {
			case "vim", "vi", "nano":
				args = []string{fmt.Sprintf("+%d", line), path}
			case "code", "cursor", "subl":
				args = []string{"--goto", fmt.Sprintf("%s:%d", path, line)}
			case "emacs":
				args = []string{fmt.Sprintf("+%d", line), path}
			default:
				args = []string{path}
			}
		} else {
			args = []string{path}
		}

		return exec.Command(editor, args...), isTerminalEditor, nil
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		opener := "xdg-open"
		if _, err := lookPath(opener); err != nil {
			return nil, false, errors.WrapError(err, "OpenWithXdgOpen")
		}
		cmd = exec.Command(opener, path)
	case "darwin":
		opener := "open"
		if _, err := lookPath(opener); err != nil {
			return nil, false, errors.WrapError(err, "OpenWithMacOSOpen")
		}
		cmd = exec.Command(opener, path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		return nil, false, exec.ErrNotFound
	}
	return cmd, false, nil
}

func isTerminalEditor(editor string) bool {
	terminalEditors := map[string]bool{
		"vim":   true,
		"nano":  true,
		"vi":    true,
		"emacs": true,
	}
	return terminalEditors[editor]
}

// IsTextFile returns true if the file extension suggests it's a text or code file.
func IsTextFile(fs core.FileSystem, path string) bool {
	ext := strings.ToLower(fs.Ext(path))
	textExts := map[string]bool{
		".txt":  true,
		".md":   true,
		".go":   true,
		".py":   true,
		".js":   true,
		".ts":   true,
		".html": true,
		".css":  true,
		".json": true,
		".yml":  true,
		".yaml": true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".hpp":  true,
		".rs":   true,
		".java": true,
		".sh":   true,
		".bash": true,
		".zsh":  true,
		".lua":  true,
		".sql":  true,
		".xml":  true,
		"":      true, // Files without extension are often text
	}
	return textExts[ext]
}
