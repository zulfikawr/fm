package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/factory"
	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/git"
	"github.com/zulfikawr/fm/internal/tui/theme"
	"golang.org/x/sync/errgroup"
)

// InfoOptions contains options for the info command
type InfoOptions struct {
	Path      string
	JSON      bool
	Tree      bool
	TreeDepth int
	Remote    string
}

// InfoResult contains the information about a path
type InfoResult struct {
	Path          string    `json:"path"`
	Type          string    `json:"type"` // "file" or "directory"
	Size          int64     `json:"size"`
	SizeFormatted string    `json:"size_formatted"`
	Permissions   string    `json:"permissions"`
	Mode          string    `json:"mode"`
	Modified      time.Time `json:"modified"`
	CanRead       bool      `json:"can_read"`
	CanWrite      bool      `json:"can_write"`

	// Directory-specific
	FileCount          int    `json:"file_count,omitempty"`
	DirectoryCount     int    `json:"directory_count,omitempty"`
	TotalSize          int64  `json:"total_size,omitempty"`
	TotalSizeFormatted string `json:"total_size_formatted,omitempty"`

	// Git-specific
	InGitRepo bool      `json:"in_git_repo"`
	GitRoot   string    `json:"git_root,omitempty"`
	GitBranch string    `json:"git_branch,omitempty"`
	GitStatus string    `json:"git_status,omitempty"`
	GitStats  *GitStats `json:"git_stats,omitempty"`

	// Tree view (only when --tree is used)
	Tree *TreeNode `json:"tree,omitempty"`
}

// GitStats contains git statistics for a directory
type GitStats struct {
	Modified  int `json:"modified"`
	Added     int `json:"added"`
	Deleted   int `json:"deleted"`
	Untracked int `json:"untracked"`
	Staged    int `json:"staged"`
}

// TreeNode represents a node in the tree view
type TreeNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"is_dir"`
	Size     int64      `json:"size"`
	Children []TreeNode `json:"children,omitempty"`
}

// RunInfo executes the info command
func RunInfo(opts InfoOptions) error {
	cfg := config.Load()
	t := theme.Themes[cfg.ThemeIndex]
	styles := theme.NewStylesheet(t)

	ctx := context.Background()

	// Initialize filesystem
	fs, _, err := factory.CreateFileSystem(opts.Remote, nil)
	if err != nil {
		return fmt.Errorf("initializing filesystem: %w", err)
	}
	defer func() {
		if closeErr := fs.Close(); closeErr != nil {
			// Log error but don't override main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close filesystem: %v\n", closeErr)
		}
	}()

	// Resolve path
	targetPath := opts.Path
	if targetPath == "" || targetPath == "." {
		if fs.IsLocal() {
			targetPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}
		} else {
			targetPath = "."
		}
	}

	// Get absolute path
	if fs.IsLocal() {
		targetPath, err = filepath.Abs(targetPath)
		if err != nil {
			return fmt.Errorf("resolving path: %w", err)
		}
	} else {
		targetPath = fs.Clean(targetPath)
	}

	// Get file info
	info, err := fs.Stat(ctx, targetPath)
	if err != nil {
		return fmt.Errorf("accessing path: %w", err)
	}

	// Initialize git service
	gs := git.NewGitService(cfg.EnableGit)

	// Build result
	result := &InfoResult{
		Path:          targetPath,
		Type:          getTypeString(info),
		Size:          info.Size(),
		SizeFormatted: format.FormatSize(info.Size(), cfg.SizeFormatIndex),
		Permissions:   info.Mode().String(),
		Mode:          fmt.Sprintf("%04o", info.Mode().Perm()),
		Modified:      info.ModTime(),
		CanRead:       canRead(info),
		CanWrite:      canWrite(info),
	}

	// Get git information
	if fs.IsLocal() && gs.IsEnabled() {
		gitRoot := gs.GetRoot(ctx, targetPath)
		if gitRoot != "" {
			result.InGitRepo = true
			result.GitRoot = gitRoot

			// Get status and branch
			statuses, branch := gs.GetStatus(ctx, targetPath)
			result.GitBranch = branch

			// If it's a file, get its status
			if !info.IsDir() {
				if status, ok := statuses[targetPath]; ok {
					result.GitStatus = status
				}
			}

			// Calculate git stats for directory
			if info.IsDir() {
				result.GitStats = calculateGitStats(statuses)
			}
		}
	}

	// Directory-specific information
	if info.IsDir() {
		if opts.Tree {
			// Build tree view
			tree, err := buildTree(ctx, fs, targetPath, 0, opts.TreeDepth)
			if err != nil {
				return fmt.Errorf("building tree: %w", err)
			}
			result.Tree = tree
		} else {
			// Calculate directory stats
			stats, err := calculateDirStats(ctx, fs, targetPath)
			if err != nil {
				return fmt.Errorf("calculating directory stats: %w", err)
			}
			result.FileCount = stats.FileCount
			result.DirectoryCount = stats.DirectoryCount
			result.TotalSize = stats.TotalSize
			result.TotalSizeFormatted = format.FormatSize(stats.TotalSize, cfg.SizeFormatIndex)
		}
	}

	// Output
	if opts.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	// Pretty print
	printInfo(result, &styles, cfg.SizeFormatIndex, opts.Tree)
	return nil
}

// DirStats contains statistics about a directory
type DirStats struct {
	FileCount      int
	DirectoryCount int
	TotalSize      int64
}

// calculateDirStats calculates statistics for a directory
func calculateDirStats(ctx context.Context, fs core.FileSystem, path string) (*DirStats, error) {
	stats := &DirStats{}

	entries, err := fs.ReadDir(ctx, path)
	if err != nil {
		return nil, err
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // Limit concurrent operations

	var mu sync.Mutex

	for _, entry := range entries {
		entry := entry
		g.Go(func() error {
			if entry.IsDir() {
				mu.Lock()
				stats.DirectoryCount++
				mu.Unlock()

				// Recursively calculate subdirectory size
				subStats, err := calculateDirStats(gctx, fs, fs.Join(path, entry.Name()))
				if err != nil {
					// Ignore permission errors
					return nil
				}
				mu.Lock()
				stats.FileCount += subStats.FileCount
				stats.DirectoryCount += subStats.DirectoryCount
				stats.TotalSize += subStats.TotalSize
				mu.Unlock()
			} else {
				mu.Lock()
				stats.FileCount++
				stats.TotalSize += entry.Size()
				mu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return stats, nil
}

// buildTree builds a tree view of the directory structure
func buildTree(ctx context.Context, fs core.FileSystem, path string, depth, maxDepth int) (*TreeNode, error) {
	info, err := fs.Stat(ctx, path)
	if err != nil {
		return nil, err
	}

	node := &TreeNode{
		Name:  info.Name(),
		Path:  path,
		IsDir: info.IsDir(),
		Size:  info.Size(),
	}

	if info.IsDir() && (maxDepth == 0 || depth < maxDepth) {
		entries, err := fs.ReadDir(ctx, path)
		if err != nil {
			// Permission error, return node without children
			return node, nil
		}

		for _, entry := range entries {
			childPath := fs.Join(path, entry.Name())
			childNode, err := buildTree(ctx, fs, childPath, depth+1, maxDepth)
			if err != nil {
				// Skip entries we can't access
				continue
			}
			node.Children = append(node.Children, *childNode)
		}
	}

	return node, nil
}

// calculateGitStats calculates git statistics from status map
func calculateGitStats(statuses map[string]string) *GitStats {
	stats := &GitStats{}

	for _, status := range statuses {
		if len(status) == 0 {
			continue
		}

		// Git status is 2 characters: XY
		// X = index status, Y = working tree status
		x := string(status[0])
		y := ""
		if len(status) > 1 {
			y = string(status[1])
		}

		// Index status
		switch x {
		case "M", "T":
			stats.Staged++
		case "A":
			stats.Added++
			stats.Staged++
		case "D":
			stats.Deleted++
			stats.Staged++
		case "R", "C":
			stats.Staged++
		}

		// Working tree status
		switch y {
		case "M", "T":
			stats.Modified++
		case "D":
			stats.Deleted++
		case "?":
			stats.Untracked++
		}
	}

	return stats
}

// printInfo prints the info result in a pretty format
func printInfo(result *InfoResult, styles *theme.Stylesheet, sizeFormat int, isTree bool) {
	// Header
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Header.GetForeground()).
		Render("File Information")
	fmt.Println(title)
	fmt.Println()

	// Basic info
	printField(styles, "Path", result.Path)
	printField(styles, "Type", result.Type)
	printField(styles, "Size", result.SizeFormatted)
	printField(styles, "Permissions", result.Permissions)
	printField(styles, "Mode", result.Mode)
	printField(styles, "Modified", result.Modified.Format("2006-01-02 15:04:05"))
	printField(styles, "Readable", boolToYesNo(result.CanRead))
	printField(styles, "Writable", boolToYesNo(result.CanWrite))

	// Directory info
	if result.Type == "directory" && !isTree {
		fmt.Println()
		printField(styles, "Files", fmt.Sprintf("%d", result.FileCount))
		printField(styles, "Directories", fmt.Sprintf("%d", result.DirectoryCount))
		printField(styles, "Total Size", result.TotalSizeFormatted)
	}

	// Git info
	if result.InGitRepo {
		fmt.Println()
		gitTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Header.GetForeground()).
			Render("Git Information")
		fmt.Println(gitTitle)
		fmt.Println()

		printField(styles, "Repository", result.GitRoot)
		printField(styles, "Branch", result.GitBranch)

		if result.GitStatus != "" {
			printField(styles, "Status", result.GitStatus)
		}

		if result.GitStats != nil {
			fmt.Println()
			printField(styles, "Modified", fmt.Sprintf("%d", result.GitStats.Modified))
			printField(styles, "Added", fmt.Sprintf("%d", result.GitStats.Added))
			printField(styles, "Deleted", fmt.Sprintf("%d", result.GitStats.Deleted))
			printField(styles, "Untracked", fmt.Sprintf("%d", result.GitStats.Untracked))
			printField(styles, "Staged", fmt.Sprintf("%d", result.GitStats.Staged))
		}
	}

	// Tree view
	if isTree && result.Tree != nil {
		fmt.Println()
		treeTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Header.GetForeground()).
			Render("Directory Tree")
		fmt.Println(treeTitle)
		fmt.Println()

		// Print root directory name
		rootName := result.Tree.Name
		if result.Tree.IsDir {
			rootName = styles.DirCol.Render(rootName + "/")
		}
		fmt.Println(rootName)

		// Print children with tree structure
		if len(result.Tree.Children) > 0 {
			for i, child := range result.Tree.Children {
				printTree(&child, "", i == len(result.Tree.Children)-1, styles, sizeFormat)
			}
		} else {
			fmt.Println(lipgloss.NewStyle().Faint(true).Render("  (empty directory)"))
		}
	}
}

// printField prints a labeled field
func printField(styles *theme.Stylesheet, label, value string) {
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.DirCol.GetForeground()).
		Width(15)

	fmt.Printf("%s %s\n", labelStyle.Render(label+":"), value)
}

// printTree prints a tree node recursively
func printTree(node *TreeNode, prefix string, isLast bool, styles *theme.Stylesheet, sizeFormat int) {
	// Determine the branch characters
	branch := "├── "
	if isLast {
		branch = "└── "
	}

	// Format name with icon
	name := node.Name
	if node.IsDir {
		name = styles.DirCol.Render(name + "/")
	} else {
		name = styles.FileCol.Render(name)
	}

	// Format size
	sizeStr := ""
	if !node.IsDir {
		sizeStr = " " + lipgloss.NewStyle().
			Faint(true).
			Render(fmt.Sprintf("(%s)", format.FormatSize(node.Size, sizeFormat)))
	}

	fmt.Printf("%s%s%s%s\n", prefix, branch, name, sizeStr)

	// Print children
	if len(node.Children) > 0 {
		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		for i, child := range node.Children {
			printTree(&child, newPrefix, i == len(node.Children)-1, styles, sizeFormat)
		}
	}
}

// Helper functions

func getTypeString(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	return "file"
}

func canRead(info os.FileInfo) bool {
	return info.Mode().Perm()&0400 != 0
}

func canWrite(info os.FileInfo) bool {
	return info.Mode().Perm()&0200 != 0
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
