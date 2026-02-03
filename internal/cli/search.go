package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/factory"
	"github.com/zulfikawr/fm/internal/files/ops"
	"github.com/zulfikawr/fm/internal/git"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// HighlightStyles defines styles for search result highlighting
type HighlightStyles struct {
	Match lipgloss.Style
	Base  lipgloss.Style
}

// RunSearch performs the fuzzy search from the CLI
func RunSearch(args *Args) error {
	cfg := config.Load()
	t := theme.Themes[cfg.ThemeIndex]
	styles := theme.NewStylesheet(t)

	if args.SearchQuery == "" {
		fmt.Println(styles.Error.Render("Error: Search query cannot be empty"))
		return nil
	}

	ctx := context.Background()

	// Determine the path to search
	searchPath := "."
	if len(args.Args) > 0 {
		searchPath = args.Args[0]
	}

	// Initialize filesystem
	fs, _, err := factory.CreateFileSystem(args.Remote, nil)
	if err != nil {
		return fmt.Errorf("initializing filesystem: %w", err)
	}
	defer func() {
		_ = fs.Close()
	}()

	if fs.IsLocal() {
		if searchPath == "." {
			searchPath, _ = os.Getwd()
		}
		searchPath, _ = fs.Abs(searchPath)
	} else {
		searchPath = fs.Clean(searchPath)
	}

	gs := git.NewGitService(true)

	// Print Search Header
	fmt.Printf("%s %s %s %s\n\n",
		styles.DirCol.Render("Searching for:"),
		styles.GitStaged.Render(args.SearchQuery),
		styles.DimCol.Render("in"),
		styles.FileCol.Render(searchPath),
	)

	results, err := ops.Search(ops.SearchOptions{
		OpCtx: ops.OpContext{Context: ctx, FS: fs},
		Git:   gs,
		Root:  searchPath,
		Query: args.SearchQuery,
	})
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println(styles.DimCol.Render("No matches found."))
		return nil
	}

	highlightStyle := lipgloss.NewStyle().
		Background(styles.SelectedItem.GetBackground()).
		Foreground(styles.Success.GetForeground()).
		Bold(true)

	for _, res := range results {
		relPath, _ := filepath.Rel(searchPath, res.Path)
		if relPath == "." || relPath == "" {
			relPath = res.Path
		}

		// Separate filename matches from content matches
		var fileNameMatch *core.Match
		for i := range res.Matches {
			if res.Matches[i].Line == 0 {
				fileNameMatch = &res.Matches[i]
				break
			}
		}

		// Render Header (Directory + Filename)
		dirPath := filepath.Dir(relPath)
		baseName := filepath.Base(relPath)

		header := " "
		if dirPath != "." {
			header += styles.DimCol.Render(dirPath + string(filepath.Separator))
		}

		if fileNameMatch != nil {
			header += highlightMatchesWithBase(baseName, fileNameMatch.MatchedIdx, HighlightStyles{
				Match: highlightStyle,
				Base:  styles.DirCol,
			})
		} else {
			header += styles.DirCol.Render(baseName)
		}

		fmt.Println(header)

		// Render Line Matches
		for _, match := range res.Matches {
			if match.Line == 0 {
				continue
			}

			linePrefix := styles.DimCol.Render(fmt.Sprintf("%5d │ ", match.Line))
			content := highlightMatches(match.Content, match.MatchedIdx, highlightStyle)
			fmt.Printf("%s%s\n", linePrefix, content)
		}
		fmt.Println()
	}

	// Summary Footer
	totalFiles := len(results)
	totalMatches := 0
	for _, res := range results {
		totalMatches += len(res.Matches)
	}

	footer := fmt.Sprintf(" Found %d matches in %d files.", totalMatches, totalFiles)
	fmt.Println(styles.Success.Render(footer))

	return nil
}

// highlightMatchesWithBase highlights specific characters in a string while applying a base style to the rest.
func highlightMatchesWithBase(content string, indices []int, styles HighlightStyles) string {
	if len(indices) == 0 {
		return styles.Base.Render(content)
	}

	isMatched := make(map[int]bool)
	for _, idx := range indices {
		isMatched[idx] = true
	}

	var sb strings.Builder
	runes := []rune(content)
	for i, r := range runes {
		char := string(r)
		if isMatched[i] {
			sb.WriteString(styles.Match.Render(char))
		} else {
			sb.WriteString(styles.Base.Render(char))
		}
	}
	return sb.String()
}

// highlightMatches highlights specific characters in a string without applying a base style to non-matches.
func highlightMatches(content string, indices []int, style lipgloss.Style) string {
	if len(indices) == 0 {
		return content
	}

	isMatched := make(map[int]bool)
	for _, idx := range indices {
		isMatched[idx] = true
	}

	var sb strings.Builder
	runes := []rune(content)
	for i, r := range runes {
		if isMatched[i] {
			sb.WriteString(style.Render(string(r)))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
