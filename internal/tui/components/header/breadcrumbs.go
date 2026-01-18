package header

import (
	"strings"

	"github.com/zulfikawr/fm/internal/tui/theme"
)

func renderBreadcrumbPath(path, separator string, remoteStr string, rootOverride string, styles theme.Stylesheet) string {
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

	if rootOverride != "" {
		// Use rootOverride for breadcrumbs (e.g. for archives)
		// We split it by the separator to style it like a path
		rootSep := "/" // Default for archives or fallback
		if strings.Contains(rootOverride, "\\") {
			rootSep = "\\"
		}

		rootParts := strings.Split(rootOverride, rootSep)
		var cleanRootParts []string
		for _, p := range rootParts {
			if p != "" {
				cleanRootParts = append(cleanRootParts, p)
			}
		}

		var styledRootParts []string
		for _, p := range cleanRootParts {
			styledRootParts = append(styledRootParts, baseStyle.Render(p))
		}

		separatorStr := dimHeaderStyle.Render(" > ")
		rootPath := strings.Join(styledRootParts, separatorStr)

		if strings.HasPrefix(rootOverride, "/") || strings.HasPrefix(rootOverride, "\\") {
			rootIndicator = baseStyle.Render(rootSep) + separatorStr + rootPath
		} else {
			rootIndicator = rootPath
		}
	} else {
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
	}

	var styledParts []string
	for _, p := range cleanParts {
		styledParts = append(styledParts, baseStyle.Render(p))
	}

	separatorStr := dimHeaderStyle.Render(" > ")
	breadcrumb := strings.Join(styledParts, separatorStr)

	if breadcrumb == "" {
		if (remoteStr != "" || rootOverride != "") && path == "/" {
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
