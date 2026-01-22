package header

import (
	"strings"

	"github.com/zulfikawr/fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// Props contains data for rendering the header
type Props struct {
	Width         int
	Path          string
	Separator     string
	RemoteStr     string
	RootOverride  string
	GitBranch     string
	ReadOnly      bool
	TabCount      int
	ActiveTab     int
	SettingsOpen  bool
	LogOpen       bool
	ClipboardOpen bool
	Style         theme.Stylesheet
}

// Render renders the application header (Breadcrumbs + Tabs)
func Render(props Props) string {
	// Account for Padding(0, 1) in styles.go
	availableWidth := props.Width - 2
	if availableWidth < 0 {
		availableWidth = 0
	}

	tabs := ""
	if shouldShowTabs(props.TabCount) {
		tabs = renderTabList(TabConfig{
			TabCount:     props.TabCount,
			ActiveIndex:  props.ActiveTab,
			ShowShortcut: true,
			Width:        props.Width,
		}, props.Style)
	}

	tabsWidth := lipgloss.Width(tabs)

	// Determine title
	title := GetTitle(TitleProps{
		Path:          props.Path,
		SettingsOpen:  props.SettingsOpen,
		LogOpen:       props.LogOpen,
		ClipboardOpen: props.ClipboardOpen,
		Style:         props.Style,
	})

	// Maximum width for breadcrumb is availableWidth - tabsWidth - gap(1)
	maxBreadcrumbWidth := availableWidth - tabsWidth
	if tabsWidth > 0 {
		maxBreadcrumbWidth -= 1 // 1 char gap
	}
	if maxBreadcrumbWidth < 0 {
		maxBreadcrumbWidth = 0
	}

	// Breadcrumb rendering
	var breadcrumb string
	if props.SettingsOpen || props.LogOpen || props.ClipboardOpen {
		breadcrumb = props.Style.Header.UnsetPadding().UnsetWidth().Render(title)
		if lipgloss.Width(breadcrumb) > maxBreadcrumbWidth {
			breadcrumb = lipgloss.NewStyle().MaxWidth(maxBreadcrumbWidth).Render(breadcrumb)
		}
	} else {
		breadcrumb = renderBreadcrumbPath(BreadcrumbProps{
			Path:         title,
			Separator:    props.Separator,
			RemoteStr:    props.RemoteStr,
			RootOverride: props.RootOverride,
			Styles:       props.Style,
			MaxWidth:     maxBreadcrumbWidth,
		})
		breadcrumb = addGitBranch(breadcrumb, props.GitBranch, props.Style)
		breadcrumb = addReadOnlyIndicator(breadcrumb, props.ReadOnly, props.Style)
	}

	breadcrumbWidth := lipgloss.Width(breadcrumb)

	gap := availableWidth - breadcrumbWidth - tabsWidth
	if gap < 0 {
		gap = 0
	}

	baseStyle := props.Style.Header.UnsetPadding().UnsetWidth()
	totalHeader := breadcrumb + baseStyle.Render(strings.Repeat(" ", gap)) + tabs
	return props.Style.Header.Width(props.Width).Render(totalHeader)
}
