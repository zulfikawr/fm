package files

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var Editors = []string{
	"vim",
	"nano",
	"vi",
	"emacs",
	"code",
	"subl",
	"cursor",
	"zed",
}

// GetOpenCmd returns the command to open a file and a boolean indicating if it's a terminal-based editor.
func GetOpenCmd(path string, editorIdx int) (*exec.Cmd, bool, error) {
	if IsTextFile(path) {
		editor := Editors[editorIdx]
		isTerminalEditor := isTerminalEditor(editor)
		return exec.Command(editor, path), isTerminalEditor, nil
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
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
func IsTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
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
