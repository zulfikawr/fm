package git

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"fm/internal/constants"
)

func (gs *gitService) GetStatus(ctx context.Context, path string) (map[string]string, string) {
	if !gs.IsEnabled() {
		return nil, ""
	}

	statuses := make(map[string]string)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, constants.GitCommandTimeout)
	defer cancel()

	repoRoot := gs.GetRoot(ctx, path)
	if repoRoot == "" {
		return statuses, ""
	}

	// Get branch
	branchCmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	branch := ""
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	// Calculate relative path
	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return statuses, branch
	}

	// Get file statuses - scope to current directory for performance, keep --ignored
	var cmd *exec.Cmd
	if relPath == "." || relPath == "" {
		// At repo root, get all statuses
		cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "status", "--porcelain", "-uall", "--ignored")
	} else {
		// In subdirectory, limit to current directory only for better performance
		cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "status", "--porcelain", "-uall", "--ignored", "--", relPath)
	}

	out, err := cmd.Output()
	if err != nil {
		return statuses, branch
	}

	statuses = ParseGitStatusPorcelain(string(out), repoRoot, path)
	return statuses, branch
}

// ParseGitStatusPorcelain is a shared helper to parse git status --porcelain output.
func ParseGitStatusPorcelain(output, repoRoot, dirPath string) map[string]string {
	statuses := make(map[string]string)
	relDir, _ := filepath.Rel(repoRoot, dirPath)
	if relDir == "." {
		relDir = ""
	}
	relDir = filepath.ToSlash(relDir)
	if relDir == "." {
		relDir = ""
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		status := line[:2]
		char := string(status[0])
		if char == " " || char == "?" || char == "!" {
			char = string(status[1])
		}
		if status == "!!" {
			char = "!"
		}

		filePath := line[3:]
		if strings.Contains(filePath, " -> ") {
			parts := strings.Split(filePath, " -> ")
			filePath = parts[1]
		}
		filePath = strings.Trim(filePath, "\"")
		filePath = filepath.ToSlash(filePath)

		if relDir != "" && !strings.HasPrefix(filePath, relDir) {
			continue
		}

		subPath := filePath
		if relDir != "" {
			subPath = strings.TrimPrefix(filePath, relDir+"/")
		}

		parts := strings.Split(subPath, "/")
		name := parts[0]

		existing := statuses[name]
		if existing == "U" {
			continue
		}
		if char == "U" || existing == "" || (existing == "!" && char != "!") || (existing == "?" && char != "?" && char != "!") || (existing == "A" && char == "M") {
			statuses[name] = char
		}
	}

	return statuses
}
