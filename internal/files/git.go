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

	// Get file statuses (including ignored)
	cmd := exec.Command("git", "-C", repoRoot, "status", "--porcelain", "--ignored")
	out, err := cmd.Output()
	if err != nil {
		return statuses, ""
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
