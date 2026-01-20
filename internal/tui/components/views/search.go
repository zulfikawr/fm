package views

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/tui/components/loading"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// SearchProps contains all data needed to render the fuzzy search results
type SearchProps struct {
	Width       int
	Height      int
	Query       string
	Results     []core.FileResult
	IsSearching bool
	CursorFile  int
	CursorMatch int
	Offset      int
	Spinner     ui.Spinner
	Style       theme.Stylesheet
}

// RenderSearch renders the fuzzy search results view
func RenderSearch(props SearchProps) string {
	if props.Height <= 0 {
		return ""
	}

	if props.Query == "" && !props.IsSearching && len(props.Results) == 0 {
		return renderSearchEmpty(props.Width, props.Height, "Type to search for content in files...", props.Style)
	}

	if props.IsSearching && len(props.Results) == 0 {
		return loading.Render(loading.Props{
			Width:   props.Width,
			Height:  props.Height,
			Message: "Searching...",
			Spinner: props.Spinner,
			Style:   props.Style,
		})
	}

	if len(props.Results) == 0 && props.Query != "" {
		return renderSearchEmpty(props.Width, props.Height, "No matches found.", props.Style)
	}

	// Stats header (Fixed)
	totalMatches := 0
	for _, res := range props.Results {
		totalMatches += len(res.Matches)
	}
	stats := fmt.Sprintf(" Found %d matches in %d files", totalMatches, len(props.Results))
	headerStr := props.Style.ListHeader.Render(stats) + "\n\n"
	headerHeight := 2

	var allLines []string

	// Render each file and its matches
	for fIdx, res := range props.Results {
		// Render file header
		prefix := "▼ "
		if res.Collapsed {
			prefix = "▶ "
		}

		// Find filename match if any
		var fileNameMatch *core.Match
		for i := range res.Matches {
			if res.Matches[i].Line == 0 {
				fileNameMatch = &res.Matches[i]
				break
			}
		}

		fileNameStr := res.FileName
		isHeaderSelected := fIdx == props.CursorFile && (res.Collapsed || props.CursorMatch == -1)

		var fileNameView string
		if fileNameMatch != nil {
			fileNameView = renderMatchContent(fileNameStr, fileNameMatch.MatchedIdx, isHeaderSelected, props.Style)
			// Apply DirCol if not selected
			if !isHeaderSelected {
				fileNameView = props.Style.DirCol.Render(fileNameView)
			}
		} else {
			if isHeaderSelected {
				fileNameView = props.Style.SelectedItem.Render(fileNameStr)
			} else {
				fileNameView = props.Style.DirCol.Render(fileNameStr)
			}
		}

		fileHeader := fmt.Sprintf("%s%s (%d)", prefix, fileNameView, len(res.Matches))
		allLines = append(allLines, fileHeader)

		if !res.Collapsed {
			// Render matches
			for mIdx, match := range props.Results[fIdx].Matches {
				if match.Line == 0 {
					continue // Skip redundant name line
				}
				isSelected := fIdx == props.CursorFile && mIdx == props.CursorMatch

				lineNum := props.Style.DimCol.Render(fmt.Sprintf("%5d: ", match.Line))
				content := renderMatchContent(match.Content, match.MatchedIdx, isSelected, props.Style)

				matchLine := "  " + lineNum + content
				allLines = append(allLines, matchLine)
			}
		}
		allLines = append(allLines, "")
	}

	// Virtualization
	resultsHeight := props.Height - headerHeight
	if resultsHeight < 0 {
		resultsHeight = 0
	}

	start := props.Offset
	if start < 0 {
		start = 0
	}
	if start >= len(allLines) && len(allLines) > 0 {
		start = len(allLines) - 1
	}

	end := start + resultsHeight
	if end > len(allLines) {
		end = len(allLines)
	}

	// Fill remaining space
	result := headerStr
	if start < end {
		result += strings.Join(allLines[start:end], "\n")
	}

	remainingRows := resultsHeight - (end - start)
	if remainingRows > 0 {
		result += strings.Repeat("\n", remainingRows)
	}

	return result
}

func renderSearchEmpty(width, height int, message string, styles theme.Stylesheet) string {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(styles.DimCol.Render(message))
}

func renderMatchContent(content string, matchedIdx []int, isSelected bool, styles theme.Stylesheet) string {
	if len(matchedIdx) == 0 {
		return content
	}

	var sb strings.Builder
	idxMap := make(map[int]bool)
	for _, idx := range matchedIdx {
		idxMap[idx] = true
	}

	// Highlight style: use theme's selected colors for the match background
	highlightStyle := lipgloss.NewStyle().
		Background(styles.SelectedItem.GetBackground()).
		Foreground(styles.Success.GetForeground()).
		Bold(true)

	// Simple match style: just the success color, no background
	matchStyle := lipgloss.NewStyle().
		Foreground(styles.Success.GetForeground()).
		Bold(true)

	runes := []rune(content)
	for i, r := range runes {
		char := string(r)
		if idxMap[i] {
			if isSelected {
				sb.WriteString(highlightStyle.Render(char))
			} else {
				sb.WriteString(matchStyle.Render(char))
			}
		} else {
			sb.WriteString(char)
		}
	}

	return sb.String()
}
