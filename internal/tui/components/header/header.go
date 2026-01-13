package header

import (
	"strings"

	"fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// Props contains all data needed to render the header
type Props struct {
	Width           int
	Path            string
	Separator       string
	GitBranch       string
	ReadOnly        bool
	TabCount        int
	ActiveTab       int
	SettingsOpen    bool
	RemoteConnected bool
	RemoteUser      string
	RemoteHost      string
	Styles          theme.Stylesheet
}

// Render renders the complete header
func Render(props Props) string {
	if props.SettingsOpen {
		return renderSettingsHeader(props)
	}
	return renderFileHeader(props)
}

// renderSettingsHeader renders the header for settings view
func renderSettingsHeader(props Props) string {
	return props.Styles.Header.Width(props.Width).Render("Settings")
}

// renderFileHeader renders the header for file browsing view
func renderFileHeader(props Props) string {
	breadcrumb := renderBreadcrumb(props)

	tabs := ""
	if props.TabCount > 1 {
		tabs = renderTabs(props)
	}

	if tabs == "" {
		return props.Styles.Header.Width(props.Width).Render(breadcrumb)
	}

	return combineHeaderElements(breadcrumb, tabs, props)
}

// renderBreadcrumb renders the breadcrumb path
func renderBreadcrumb(props Props) string {
	remoteStr := ""
	if props.RemoteConnected {
		// Use the existing RenderRemote logic but get the raw string if we want it in breadcrumb
		remoteStr = props.RemoteUser + "@" + props.RemoteHost
		if props.RemoteUser == "" {
			remoteStr = props.RemoteHost
		}
	}

	breadcrumb := renderBreadcrumbPath(props.Path, props.Separator, remoteStr, props.Styles)
	breadcrumb = addGitBranch(breadcrumb, props.GitBranch, props.Styles)
	breadcrumb = addReadOnlyIndicator(breadcrumb, props.ReadOnly, props.Styles)
	return breadcrumb
}

// renderTabs renders the tab indicators
func renderTabs(props Props) string {
	config := TabConfig{
		TabCount:     props.TabCount,
		ActiveIndex:  props.ActiveTab,
		ShowShortcut: true,
	}
	return renderTabList(config, props.Styles)
}

// combineHeaderElements combines breadcrumb and tabs with proper spacing
func combineHeaderElements(breadcrumb, tabs string, props Props) string {
	baseHeaderStyle := props.Styles.Header.UnsetPadding().UnsetWidth()

	breadcrumbWidth := lipgloss.Width(breadcrumb)
	tabsWidth := lipgloss.Width(tabs)

	gap := props.Width - breadcrumbWidth - tabsWidth - 2 // -2 for padding
	if gap < 1 {
		gap = 1
	}

	fullHeader := breadcrumb + baseHeaderStyle.Render(strings.Repeat(" ", gap)) + tabs
	return props.Styles.Header.Width(props.Width).Render(fullHeader)
}
