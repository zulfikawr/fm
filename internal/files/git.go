package files

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GetGitStatus returns a map of filenames to their Git status relative to the given directory.
func GetGitStatus(dirPath string) (map[string]string, string) {
	statuses := make(map[string]string)

	// Get repo root to resolve relative paths correctly
	rootCmd := exec.Command("git", "-C", dirPath, "rev-parse", "--show-toplevel")
	rootOut, err := rootCmd.Output()
	if err != nil {
		return statuses, ""
	}
	repoRoot := strings.TrimSpace(string(rootOut))

	// Get branch info
	branchCmd := exec.Command("git", "-C", dirPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, _ := branchCmd.Output()
	branch := strings.TrimSpace(string(branchOut))

	// Get file statuses
	cmd := exec.Command("git", "-C", repoRoot, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return statuses, ""
	}

	// Calculate current directory relative to repo root
	relDir, _ := filepath.Rel(repoRoot, dirPath)
	if relDir == "." {
		relDir = ""
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		status := line[:2]
		// Porcelain format: XY PATH
		// X: Index status, Y: Working tree status
		char := string(status[0])
		if char == " " || char == "?" {
			char = string(status[1])
		}

		filePath := line[3:]
		// Handle renamed files: "R  old -> new"
		if strings.Contains(filePath, " -> ") {
			parts := strings.Split(filePath, " -> ")
			filePath = parts[1]
		}
		// Unquote if necessary
		filePath = strings.Trim(filePath, "\"")

		// We only care about files inside or equal to the current relDir
		if relDir != "" && !strings.HasPrefix(filePath, relDir) {
			continue
		}

		// Get the part of the path immediately under relDir
		subPath := filePath
		if relDir != "" {
			subPath = strings.TrimPrefix(filePath, relDir+"/")
		}

		parts := strings.Split(subPath, "/")
		name := parts[0]

		// Priority for display: Conflict (U) > Modified (M) > Staged (A) > Untracked (?)
		// If we already have a higher priority status for this name (dir), don't overwrite
		existing := statuses[name]
		if existing == "U" {
			continue
		}
		if char == "U" || existing == "" || (existing == "?" && char != "?") || (existing == "A" && char == "M") {
			statuses[name] = char
		}
	}

	return statuses, branch
}
