package header

import (
	"strings"

	"fm/internal/tui/theme"
)

// renderBreadcrumbPath renders the breadcrumb navigation path
func renderBreadcrumbPath(path, separator string, styles theme.Stylesheet) string {
	sep := separator
	parts := strings.Split(path, sep)
	var cleanParts []string

	if path == sep {
		cleanParts = []string{sep}
	} else {
		for _, p := range parts {
			if p != "" {
				cleanParts = append(cleanParts, p)
			}
		}
	}

	dimHeaderStyle := styles.DimCol.Inherit(styles.Header)

	var styledParts []string
	for _, p := range cleanParts {
		styledParts = append(styledParts, styles.Header.UnsetPadding().UnsetWidth().Render(p))
	}

	separatorStr := dimHeaderStyle.Render(" > ")
	breadcrumb := strings.Join(styledParts, separatorStr)

	if strings.HasPrefix(path, sep) && path != sep {
		rootStyle := styles.Header.UnsetPadding().UnsetWidth()
		breadcrumb = rootStyle.Render(sep) + separatorStr + breadcrumb
	}

	return breadcrumb
}

// addGitBranch adds git branch indicator to breadcrumb if present
func addGitBranch(breadcrumb, gitBranch string, styles theme.Stylesheet) string {
	if gitBranch == "" {
		return breadcrumb
	}

	gitStyle := styles.GitStaged.Inherit(styles.Header).UnsetPadding().UnsetWidth()
	dimStyle := styles.DimCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()

	gitIndicator := dimStyle.Render(" (") + gitStyle.Render(gitBranch) + dimStyle.Render(")")
	return breadcrumb + gitIndicator
}

// addReadOnlyIndicator adds read-only indicator if applicable
func addReadOnlyIndicator(breadcrumb string, readOnly bool, styles theme.Stylesheet) string {
	if !readOnly {
		return breadcrumb
	}

	dimStyle := styles.DimCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()
	roStyle := styles.KeyCol.Inherit(styles.Header).UnsetPadding().UnsetWidth()

	roIndicator := dimStyle.Render(" [") + roStyle.Render("RO") + dimStyle.Render("]")
	return breadcrumb + roIndicator
}
