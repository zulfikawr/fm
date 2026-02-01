package header

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// BreadcrumbProps encapsulates data for rendering breadcrumbs
type BreadcrumbProps struct {
	Path         string
	Separator    string
	RemoteStr    string
	RootOverride string
	Styles       theme.Stylesheet
	MaxWidth     int
}

func renderBreadcrumbPath(props BreadcrumbProps) string {
	sep := props.Separator
	parts := strings.Split(props.Path, sep)
	var cleanParts []string

	for _, p := range parts {
		if p != "" {
			cleanParts = append(cleanParts, p)
		}
	}

	dimHeaderStyle := props.Styles.MutedCol.Inherit(props.Styles.Header)
	baseStyle := props.Styles.Header.UnsetPadding().UnsetWidth()

	// Determine the root indicator (local sep or remote string)
	var rootIndicator string
	rootIndicatorRaw := sep

	if props.RootOverride != "" {
		// Use rootOverride for breadcrumbs (e.g. for archives)
		rootIndicatorRaw = props.RootOverride
	} else {
		// Handle Windows drive letters (e.g., C:)
		if len(cleanParts) > 0 && strings.Contains(cleanParts[0], ":") && sep == "\\" {
			rootIndicatorRaw = cleanParts[0]
			cleanParts = cleanParts[1:]
		}

		if props.RemoteStr != "" {
			rootIndicatorRaw = strings.TrimSuffix(props.RemoteStr, "/")
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

	if props.MaxWidth <= 0 || lipgloss.Width(fullBreadcrumb) <= props.MaxWidth {
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
		if lipgloss.Width(collapsed) <= props.MaxWidth {
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
		if lipgloss.Width(collapsed) <= props.MaxWidth {
			return collapsed
		}
	}

	// If still too long, just the last part or root
	if len(cleanParts) > 0 {
		lastPart := baseStyle.Render(cleanParts[len(cleanParts)-1])
		if lipgloss.Width(lastPart) <= props.MaxWidth {
			return lastPart
		}
	}

	return lipgloss.NewStyle().MaxWidth(props.MaxWidth).Render(fullBreadcrumb)
}

func addGitBranch(breadcrumb, gitBranch string, styles theme.Stylesheet) string {
	if gitBranch == "" {
		return breadcrumb
	}

	gitStyle := styles.AccentCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()
	dimStyle := styles.MutedCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()

	gitIndicator := dimStyle.Render(" (") + gitStyle.Render(gitBranch) + dimStyle.Render("*)")
	return breadcrumb + gitIndicator
}

func addReadOnlyIndicator(breadcrumb string, readOnly bool, styles theme.Stylesheet) string {
	if !readOnly {
		return breadcrumb
	}

	dimStyle := styles.MutedCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()
	roStyle := styles.InfoCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()

	roIndicator := dimStyle.Render(" [") + roStyle.Render("RO") + dimStyle.Render("]")
	return breadcrumb + roIndicator
}
