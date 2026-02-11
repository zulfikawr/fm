package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/files"
	"github.com/zulfikawr/fm/internal/files/factory"
	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/logger"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// RunAnalyze executes the disk usage analysis from CLI
func RunAnalyze(args *Args) error {
	cfg := config.Load()
	t := theme.Themes[cfg.UI.ThemeIndex]
	styles := theme.NewStylesheet(t)

	ctx := context.Background()

	// Initialize filesystem
	fs, fsInfo, err := factory.CreateFileSystem(args.Remote, nil)
	if err != nil {
		return fmt.Errorf("initializing filesystem: %w", err)
	}
	if fsInfo != nil {
		logger.Debugf("Connected to remote for analysis: %s@%s", fsInfo.User, fsInfo.Host)
	}
	defer logger.CloseAndLog(fs, "analyze filesystem")

	// Resolve path
	targetPath := "."
	if len(args.Args) > 0 {
		targetPath = args.Args[0]
	}

	if fs.IsLocal() {
		if targetPath == "." {
			if cwd, err := os.Getwd(); err == nil {
				targetPath = cwd
			}
		}
		if absPath, err := fs.Abs(targetPath); err == nil {
			targetPath = absPath
		}
	} else {
		targetPath = fs.Clean(targetPath)
	}

	fmt.Printf("%s %s...\n", styles.InfoCol.Render("Analyzing"), targetPath)

	analyzer := files.NewAnalyzer(fs)
	result, err := analyzer.AnalyzeConcurrent(ctx, targetPath, nil)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if result == nil {
		return fmt.Errorf("no results found")
	}

	// Print Summary
	fmt.Println()
	title := lipgloss.NewStyle().Bold(true).Foreground(styles.Header.GetForeground()).Render("Disk Usage Analysis")
	fmt.Println(title)
	fmt.Println(styles.DimCol.Render(targetPath))
	fmt.Println(strings.Repeat("─", 65))

	// Constants for alignment
	const (
		nameWidth = 30
		sizeWidth = 10
		percWidth = 7
		barWidth  = 10
	)

	// Print immediate children sorted by size
	for i := range result.Children {
		child := result.Children[i]
		// 1. Name Column
		displayName := child.Name
		if child.IsDirectory {
			displayName += "/"
		}
		if lipgloss.Width(displayName) > nameWidth {
			displayName = displayName[:nameWidth-3] + "..."
		}

		var styledName string
		if child.IsDirectory {
			styledName = styles.DirCol.Render(displayName)
		} else {
			styledName = styles.FileCol.Render(displayName)
		}
		nameCell := lipgloss.NewStyle().Width(nameWidth).Render(styledName)

		// 2. Bar Column
		bar := renderCLIBar(child.Percentage, barWidth, &styles)

		// 3. Size Column
		sizeStr := format.FormatSize(child.Size, cfg.UI.SizeFormatIndex)
		sizeCell := styles.AccentCol.Width(sizeWidth).Align(lipgloss.Right).Render(sizeStr)

		// 4. Percentage Column
		percStr := fmt.Sprintf("%5.1f%%", child.Percentage*100)
		percCell := styles.MutedCol.Width(percWidth).Align(lipgloss.Right).Render(percStr)

		// Print the final aligned row
		fmt.Printf("%s  %s  %s  %s\n", nameCell, bar, sizeCell, percCell)
	}

	return nil
}

func renderCLIBar(percent float64, width int, styles *theme.Stylesheet) string {
	filled := int(float64(width) * percent)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := styles.PrimaryCol.Render(strings.Repeat("#", filled)) +
		styles.MutedCol.Render(strings.Repeat(".", width-filled))

	return "[" + bar + "]"
}
