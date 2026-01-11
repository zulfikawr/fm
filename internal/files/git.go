package files

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const GitCommandTimeout = 10 * time.Second

// GetGitRoot returns the root directory of the Git repository containing the given path.
func GetGitRoot(ctx context.Context, dirPath string) string {
	ctx, cancel := context.WithTimeout(ctx, GitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", dirPath, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetGitStatus returns a map of filenames to their Git status relative to the given directory.
// Optimized for performance: combines git commands, limits scope to current directory.
func GetGitStatus(ctx context.Context, dirPath string) (map[string]string, string) {
	statuses := make(map[string]string)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, GitCommandTimeout)
	defer cancel()

	// Get repo root and branch in one command for efficiency
	rootCmd := exec.CommandContext(ctx, "git", "-C", dirPath, "rev-parse", "--show-toplevel", "--abbrev-ref", "HEAD")
	rootOut, err := rootCmd.Output()
	if err != nil {
		return statuses, ""
	}

	// Parse both repo root and branch from output
	lines := strings.Split(strings.TrimSpace(string(rootOut)), "\n")
	if len(lines) < 2 {
		return statuses, ""
	}
	repoRoot := strings.TrimSpace(lines[0])
	branch := strings.TrimSpace(lines[1])

	// Calculate relative path
	relPath, err := filepath.Rel(repoRoot, dirPath)
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

	statuses = ParseGitStatusPorcelain(string(out), repoRoot, dirPath)
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
