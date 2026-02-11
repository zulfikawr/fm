package git

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
)

func (gs *gitService) GetStatus(ctx context.Context, path string) (map[string]string, string, int, int, int) {
	if !gs.IsEnabled() {
		return nil, "", 0, 0, 0
	}

	statuses := make(map[string]string)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, constants.GitCommandTimeout)
	defer cancel()

	repoRoot := gs.GetRoot(ctx, path)
	if repoRoot == "" {
		return statuses, "", 0, 0, 0
	}

	// Get branch
	branchCmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	branch := ""
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	// Get FULL repository status for stats
	fullCmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "status", "--porcelain", "-uall")
	fullOut, err := fullCmd.Output()
	var modified, staged, untracked int
	if err == nil {
		modified, staged, untracked = ParseGitStatusStats(string(fullOut))
	}

	// Calculate relative path for scoped status
	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return statuses, branch, modified, staged, untracked
	}

	// Get file statuses scoped to current directory for visual markers
	var cmd *exec.Cmd
	if relPath == "." || relPath == "" {
		cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "status", "--porcelain", "-uall", "--ignored")
	} else {
		cmd = exec.CommandContext(ctx, "git", "-C", repoRoot, "status", "--porcelain", "-uall", "--ignored", "--", relPath)
	}

	out, err := cmd.Output()
	if err != nil {
		return statuses, branch, modified, staged, untracked
	}

	statuses = ParseGitStatusPorcelain(string(out), repoRoot, path)
	return statuses, branch, modified, staged, untracked
}

// ParseGitStatusStats calculates summary statistics for the whole repo
func ParseGitStatusStats(output string) (modified, staged, untracked int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		x := status[0] // Index
		y := status[1] // Working Tree

		// Porcelain format XY:
		// ?? = untracked
		// !! = ignored
		// ' ' = unmodified
		// M = modified
		// A = added
		// D = deleted
		// R = renamed
		// C = copied
		// U = updated but unmerged

		// 1. Untracked
		if x == '?' {
			untracked++
			continue
		}

		// 2. Staged (Index column x is not empty and not untracked/ignored)
		if x != ' ' && x != '?' && x != '!' {
			staged++
		}

		// 3. Modified (Working tree column y is not empty)
		// Logic: If there is a status in column y, it represents a change in the working tree.
		// For the summary header, we count any working tree change as "Modified".
		if y != ' ' && y != '?' && y != '!' {
			modified++
		}
	}
	return
}

func (gs *gitService) GetIgnoredFiles(ctx context.Context, repoRoot string) ([]string, error) {
	if !gs.IsEnabled() {
		return nil, nil
	}

	cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "ls-files", "--others", "--ignored", "--exclude-standard", "--directory")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, f := range files {
		if f != "" {
			result = append(result, filepath.Join(repoRoot, f))
		}
	}
	return result, nil
}

// ParseGitStatusPorcelain is a shared helper to parse git status --porcelain output.
func ParseGitStatusPorcelain(output, repoRoot, dirPath string) map[string]string {
	statuses := make(map[string]string)
	relDir, err := filepath.Rel(repoRoot, dirPath)
	if err != nil {
		relDir = ""
	}
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
