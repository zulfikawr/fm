package header

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

func renderBreadcrumbPath(path, separator string, remoteStr string, rootOverride string, styles theme.Stylesheet, maxWidth int) string {
	sep := separator
	parts := strings.Split(path, sep)
	var cleanParts []string

	for _, p := range parts {
		if p != "" {
			cleanParts = append(cleanParts, p)
		}
	}

	dimHeaderStyle := styles.DimCol.Inherit(styles.Header)
	baseStyle := styles.Header.UnsetPadding().UnsetWidth()

	// Determine the root indicator (local sep or remote string)
	var rootIndicator string
	rootIndicatorRaw := sep

	if rootOverride != "" {
		// Use rootOverride for breadcrumbs (e.g. for archives)
		rootIndicatorRaw = rootOverride
	} else {
		// Handle Windows drive letters (e.g., C:)
		if len(cleanParts) > 0 && strings.Contains(cleanParts[0], ":") && sep == "\\" {
			rootIndicatorRaw = cleanParts[0]
			cleanParts = cleanParts[1:]
		}

		if remoteStr != "" {
			rootIndicatorRaw = strings.TrimSuffix(remoteStr, "/")
		}
	}

	rootIndicator = baseStyle.Render(rootIndicatorRaw)

	// Build the components
	var styledParts []string
	for _, p := range cleanParts {
		styledParts = append(styledParts, baseStyle.Render(p))
	}

	separatorStr := dimHeaderStyle.Render(" > ")

	// If it fits, return the full breadcrumb
	fullBreadcrumb := rootIndicator
	if len(styledParts) > 0 {
		fullBreadcrumb += separatorStr + strings.Join(styledParts, separatorStr)
	}

	if maxWidth <= 0 || lipgloss.Width(fullBreadcrumb) <= maxWidth {
		return fullBreadcrumb
	}

	// Collapse if too long
	// Keep root, first part (if any), last part, and use ... in between
	if len(cleanParts) > 2 {
		collapsedParts := []string{
			baseStyle.Render(cleanParts[0]),
			baseStyle.Render("..."),
			baseStyle.Render(cleanParts[len(cleanParts)-1]),
		}
		collapsed := rootIndicator + separatorStr + strings.Join(collapsedParts, separatorStr)
		if lipgloss.Width(collapsed) <= maxWidth {
			return collapsed
		}
	}

	// If still too long, just root and last part
	if len(cleanParts) > 1 {
		collapsedParts := []string{
			baseStyle.Render("..."),
			baseStyle.Render(cleanParts[len(cleanParts)-1]),
		}
		collapsed := rootIndicator + separatorStr + strings.Join(collapsedParts, separatorStr)
		if lipgloss.Width(collapsed) <= maxWidth {
			return collapsed
		}
	}

	// If still too long, just the last part or root
	if len(cleanParts) > 0 {
		lastPart := baseStyle.Render(cleanParts[len(cleanParts)-1])
		if lipgloss.Width(lastPart) <= maxWidth {
			return lastPart
		}
	}

	return lipgloss.NewStyle().MaxWidth(maxWidth).Render(fullBreadcrumb)
}

func addGitBranch(breadcrumb, gitBranch string, styles theme.Stylesheet) string {
	if gitBranch == "" {
		return breadcrumb
	}

	gitStyle := styles.GitStaged.Inherit(styles.Header).UnsetPadding().UnsetWidth()
	dimStyle := styles.DimCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()

	gitIndicator := dimStyle.Render(" (") + gitStyle.Render(gitBranch) + dimStyle.Render("*)")
	return breadcrumb + gitIndicator
}

func addReadOnlyIndicator(breadcrumb string, readOnly bool, styles theme.Stylesheet) string {
	if !readOnly {
		return breadcrumb
	}

	dimStyle := styles.DimCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()
	roStyle := styles.KeyCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()

	roIndicator := dimStyle.Render(" [") + roStyle.Render("RO") + dimStyle.Render("]")
	return breadcrumb + roIndicator
}
