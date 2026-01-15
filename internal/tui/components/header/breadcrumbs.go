package header

import (
	"strings"

	"fm/internal/tui/theme"
)

func renderBreadcrumbPath(path, separator string, remoteStr string, styles theme.Stylesheet) string {
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
	rootIndicator := baseStyle.Render(sep)

	// Handle Windows drive letters (e.g., C:)
	if len(cleanParts) > 0 && strings.Contains(cleanParts[0], ":") && sep == "\\" {
		rootIndicator = baseStyle.Render(cleanParts[0])
		cleanParts = cleanParts[1:]
	}

	if remoteStr != "" {
		// Ensure remoteStr doesn't have a trailing slash if we're going to join it
		r := strings.TrimSuffix(remoteStr, "/")
		rootIndicator = baseStyle.Render(r)
	}

	var styledParts []string
	for _, p := range cleanParts {
		styledParts = append(styledParts, baseStyle.Render(p))
	}

	separatorStr := dimHeaderStyle.Render(" > ")
	breadcrumb := strings.Join(styledParts, separatorStr)

	if breadcrumb == "" {
		if remoteStr != "" && path == "/" {
			return rootIndicator
		}
		// If path is just "/" or empty and we have a root indicator
		return rootIndicator
	}

	return rootIndicator + separatorStr + breadcrumb
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
