package ops

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
)

var lookPath = exec.LookPath

// GetOpenCmd returns the command to open a file based on the environment and configuration.
func GetOpenCmd(opts OpenOptions) (*exec.Cmd, bool, error) {
	if !opts.FS.IsLocal() {
		return nil, false, fmt.Errorf("remote open not supported")
	}
	return GetOpenAtLineCmd(opts)
}

// GetOpenAtLineCmd returns the command to open a file at a specific line.
func GetOpenAtLineCmd(opts OpenOptions) (*exec.Cmd, bool, error) {
	if !opts.FS.IsLocal() {
		return nil, false, fmt.Errorf("remote open not supported")
	}

	editor := constants.Editors[opts.EditorIdx]
	isTermEditor := isTerminalEditor(editor)

	// Check if editor exists
	if path, err := lookPath(editor); err != nil {
		return nil, false, errors.WrapError(err, fmt.Sprintf("OpenWithEditor (found at: %q)", path))
	}

	var args []string
	if opts.Line > 0 {
		switch editor {
		case "vim", "vi", "nano":
			args = []string{fmt.Sprintf("+%d", opts.Line), opts.Path}
		case "code", "cursor", "subl":
			args = []string{"--goto", fmt.Sprintf("%s:%d", opts.Path, opts.Line)}
		case "emacs":
			args = []string{fmt.Sprintf("+%d", opts.Line), opts.Path}
		default:
			args = []string{opts.Path}
		}
	} else {
		args = []string{opts.Path}
	}

	return exec.Command(editor, args...), isTermEditor, nil
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
