package ops

import (
	"context"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
)

// GetPathCompletions returns a list of possible completions for a given path prefix
func GetPathCompletions(ctx context.Context, fs core.FileSystem, currentDir, input string) []string {
	if input == "" {
		return nil
	}

	searchDir := currentDir
	prefix := input

	// Handle home directory
	if strings.HasPrefix(input, "~") {
		home, err := fs.GetHomeDir()
		if err == nil {
			input = strings.Replace(input, "~", home, 1)
		}
	}

	// Determine the directory to search in
	if strings.Contains(input, fs.Separator()) {
		searchDir = fs.Dir(input)
		prefix = fs.Base(input)
		// If input ends with separator, we are searching inside that directory
		if strings.HasSuffix(input, fs.Separator()) {
			searchDir = input
			prefix = ""
		}
	} else if strings.HasPrefix(input, "/") || (fs.Separator() == "\\" && len(input) >= 2 && input[1] == ':') {
		// Absolute path but no separator after drive or root
		searchDir = input
		if !strings.HasSuffix(input, fs.Separator()) {
			searchDir = fs.Dir(input)
			prefix = fs.Base(input)
		}
	} else {
		prefix = input
	}

	entries, err := fs.ReadDir(ctx, searchDir)
	if err != nil {
		return nil
	}

	var completions []string
	prefixLower := strings.ToLower(prefix)
	for i := range entries {
		entry := entries[i]
		name := entry.Name()
		if strings.HasPrefix(strings.ToLower(name), prefixLower) {
			completion := fs.Join(searchDir, name)
			if entry.IsDir() {
				completion += fs.Separator()
			}
			completions = append(completions, completion)
		}
	}

	return completions
}
