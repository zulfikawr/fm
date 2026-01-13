package list

import "fm/internal/files/core"

// renderGitMarker renders the git status indicator (e.g., M, A, ?)
func renderGitMarker(props Props, item core.Item, isCursor bool) string {
	gitMarker := "  "
	if item.GitStatus != "" {
		gitMarker = item.GitStatus + " "
	}

	if isCursor {
		sStyle := props.Styles.SelectedItem.UnsetPadding().UnsetWidth()
		if style, ok := props.Styles.GitStyles[item.GitStatus]; ok {
			return style.Inherit(sStyle).Render(gitMarker)
		}
		return sStyle.Render(gitMarker)
	}

	if style, ok := props.Styles.GitStyles[item.GitStatus]; ok {
		return style.Render(gitMarker)
	}
	return gitMarker
}
