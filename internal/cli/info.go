package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/files"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/factory"
	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/git"
	"github.com/zulfikawr/fm/internal/tui/theme"
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
	t := theme.Themes[cfg.UI.ThemeIndex]
	styles := theme.NewStylesheet(t)

	ctx := context.Background()

	// Initialize filesystem
	fs, fsInfo, err := factory.CreateFileSystem(opts.Remote, nil)
	if err != nil {
		return fmt.Errorf("initializing filesystem: %w", err)
	}
	if fsInfo != nil {
		// Log or handle remote info
	}
	defer func() {
		if closeErr := fs.Close(); closeErr != nil {
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

	// Build result
	result := &InfoResult{
		Path:          targetPath,
		Type:          getTypeString(info),
		Size:          info.Size(),
		SizeFormatted: format.FormatSize(info.Size(), cfg.UI.SizeFormatIndex),
		Permissions:   info.Mode().String(),
		Mode:          fmt.Sprintf("%04o", info.Mode().Perm()),
		Modified:      info.ModTime(),
		CanRead:       canRead(info),
		CanWrite:      canWrite(info),
	}

	// Initialize git service
	gs := git.NewGitService(cfg.External.EnableGit)

	// Get git information
	if fs.IsLocal() && gs.IsEnabled() {
		gitRoot := gs.GetRoot(ctx, targetPath)
		if gitRoot != "" {
			result.InGitRepo = true
			result.GitRoot = gitRoot

			// Get status and branch
			statuses, branch, modified, staged, untracked := gs.GetStatus(ctx, targetPath)
			result.GitBranch = branch

			if !info.IsDir() {
				if status, ok := statuses[targetPath]; ok {
					result.GitStatus = status
				}
			}

			if info.IsDir() {
				result.GitStats = &GitStats{
					Modified:  modified,
					Staged:    staged,
					Untracked: untracked,
				}
			}
		}
	}

	// Directory-specific information
	if info.IsDir() {
		// Use the Analyzer for accurate recursive stats
		analyzer := files.NewAnalyzer(fs)
		analysis, err := analyzer.AnalyzeConcurrent(ctx, targetPath, nil)
		if err != nil {
			return fmt.Errorf("analyzing directory: %w", err)
		}

		if analysis != nil {
			fc, dc := countRecursive(analysis)
			result.FileCount = fc
			result.DirectoryCount = dc
			result.TotalSize = analysis.Size
			result.TotalSizeFormatted = format.FormatSize(analysis.Size, cfg.UI.SizeFormatIndex)

			// Overwrite the base size with recursive size for directories
			result.SizeFormatted = result.TotalSizeFormatted
		}

		if opts.Tree {
			result.Tree = buildTreeFromAnalysis(analysis, 0, opts.TreeDepth)
		}
	}

	// Output
	if opts.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	// Pretty print
	printInfo(result, &styles, cfg.UI.SizeFormatIndex, opts.Tree)
	return nil
}

func countRecursive(node *core.AnalysisResult) (files, dirs int) {
	for i := range node.Children {
		child := node.Children[i]
		if child.IsDirectory {
			dirs++
			cf, cd := countRecursive(child)
			files += cf
			dirs += cd
		} else {
			files++
		}
	}
	return
}

func buildTreeFromAnalysis(node *core.AnalysisResult, depth, maxDepth int) *TreeNode {
	if node == nil {
		return nil
	}

	treeNode := &TreeNode{
		Name:  node.Name,
		Path:  node.Path,
		IsDir: node.IsDirectory,
		Size:  node.Size,
	}

	if node.IsDirectory && (maxDepth == 0 || depth < maxDepth) {
		for i := range node.Children {
			treeNode.Children = append(treeNode.Children, *buildTreeFromAnalysis(node.Children[i], depth+1, maxDepth))
		}
	}

	return treeNode
}

// printInfo prints the info result in a pretty format
func printInfo(result *InfoResult, styles *theme.Stylesheet, sizeFormat int, isTree bool) {
	// Header
	fmt.Println(styles.DirCol.Render("File Information"))
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
		// Total size is already shown in the "Size" field above for directories
	}

	// Git info
	if result.InGitRepo {
		fmt.Println()
		fmt.Println(styles.DirCol.Render("Git Information"))
		fmt.Println()

		printField(styles, "Repository", result.GitRoot)
		printField(styles, "Branch", result.GitBranch)

		if result.GitStatus != "" {
			printField(styles, "Status", result.GitStatus)
		}

		if result.GitStats != nil {
			fmt.Println()
			printField(styles, "Staged", fmt.Sprintf("%d", result.GitStats.Staged))
			printField(styles, "Modified", fmt.Sprintf("%d", result.GitStats.Modified))
			printField(styles, "Untracked", fmt.Sprintf("%d", result.GitStats.Untracked))
		}
	}

	// Tree view
	if isTree && result.Tree != nil {
		fmt.Println()
		fmt.Println(styles.DirCol.Render("Directory Tree"))
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
	labelStyle := styles.GitStaged.Width(15).Bold(true)
	valueStyle := styles.FileCol

	// Special values
	switch label {
	case "Path":
		valueStyle = styles.AccentCol
	case "Branch":
		valueStyle = styles.HighlightCol
	case "Size", "Total Size":
		valueStyle = styles.HighlightCol
	}

	fmt.Printf("%s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
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
		sizeStr = " " + styles.DimCol.Render(fmt.Sprintf("(%s)", format.FormatSize(node.Size, sizeFormat)))
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
