package views

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/format"
	"github.com/zulfikawr/fm/internal/tui/components/messages"
	"github.com/zulfikawr/fm/internal/tui/components/ui"
	"github.com/zulfikawr/fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// SettingItem represents a single setting or keybinding row
type SettingItem struct {
	Label        string
	Value        string
	Inactive     bool
	IsKeybinding bool   // Indicates if this is a keybinding setting
	Action       string // The action this keybinding controls
}

// SettingGroup represents a categorized group of settings
type SettingGroup struct {
	Title    string
	Settings []SettingItem
}

// SettingsProps contains data for rendering the settings view
type SettingsProps struct {
	Width  int
	Height int
	Cursor int
	Offset int
	Config config.Config
	Style  theme.Stylesheet
}

// RenderSettings renders the settings menu view
func RenderSettings(props SettingsProps) string {
	if props.Height <= 0 {
		return ""
	}

	groups := buildSettingGroups(props)

	rows := renderGroups(props, groups)

	// Apply scroll offset
	if props.Offset > 0 && props.Offset < len(rows) {
		rows = rows[props.Offset:]
	} else if props.Offset >= len(rows) {
		rows = []string{}
	}

	// Ensure we fill the viewport
	if len(rows) > props.Height {
		rows = rows[:props.Height]
	} else {
		for i := len(rows); i < props.Height; i++ {
			rows = append(rows, "")
		}
	}

	return strings.Join(rows, "\n")
}

func buildSettingGroups(props SettingsProps) []SettingGroup {
	s := props.Config
	styles := props.Style

	groups := []SettingGroup{
		{
			Title: "File Operations",
			Settings: []SettingItem{
				{Label: "Show Hidden Files", Value: ui.Toggle(s.UI.ShowHidden, styles)},
				{Label: "Case-Sensitive Search", Value: ui.Toggle(s.Ops.CaseSensitive, styles)},
				{Label: "Confirm Operations", Value: ui.Toggle(s.Ops.ConfirmOperations, styles)},
				{Label: "Wrap Navigation", Value: ui.Toggle(s.Ops.WrapNavigation, styles)},
				{Label: "Preferred Editor", Value: ui.Picker(constants.Editors[s.External.EditorIndex], styles)},
				{Label: "Use Trash (Move to Trash)", Value: ui.Toggle(s.Trash.UseTrash, styles)},
			},
		},
		{
			Title: "Display Options",
			Settings: []SettingItem{
				{Label: "Show Column Headers", Value: ui.Toggle(s.UI.ShowHeader, styles)},
				{Label: "Show RAM Usage", Value: ui.Toggle(s.UI.ShowRAMUsage, styles)},
				{Label: "Enable Git Status", Value: ui.Toggle(s.External.EnableGit, styles)},
				{Label: "Show File Size", Value: ui.Toggle(s.UI.ShowSize, styles)},
				{Label: "Size Format", Value: ui.Picker(format.SizeFormats[s.UI.SizeFormatIndex], styles)},
				{Label: "Show Date Modified", Value: ui.Toggle(s.UI.ShowDateModified, styles)},
				{Label: "Date Format", Value: ui.Picker(format.DateFormats[s.UI.DateFormatIndex].Name, styles)},
				{Label: "Enable Mouse Support", Value: ui.Toggle(s.UI.EnableMouse, styles)},
			},
		},
		{
			Title: "Search, Filtering & Inputs",
			Settings: []SettingItem{
				{Label: "Enable Regex Search", Value: ui.Toggle(s.Ops.EnableRegexSearch, styles)},
			},
		},
		{
			Title: "Appearance",
			Settings: []SettingItem{
				{Label: "Enable Nerd Font Icons", Value: ui.Toggle(s.UI.EnableIcons, styles)},
				{Label: "Theme", Value: ui.Picker(theme.Themes[s.UI.ThemeIndex].Name, styles)},
			},
		},
	}

	// Add Keybindings group
	categories := []struct {
		ID    string
		Title string
	}{
		{"navigation", "Keybindings: Navigation"},
		{"file_ops", "Keybindings: File Operations"},
		{"tabs", "Keybindings: Tabs"},
		{"selection", "Keybindings: Selection"},
		{"search", "Keybindings: Search & Filter"},
		{"general", "Keybindings: General"},
	}

	for _, cat := range categories {
		group := SettingGroup{Title: cat.Title}
		for _, kb := range s.Keybindings {
			if kb.Category == cat.ID {
				displayKeys := make([]string, len(kb.Keys))
				for i, k := range kb.Keys {
					if k == " " {
						displayKeys[i] = "[space]"
					} else {
						displayKeys[i] = "[" + k + "]"
					}
				}
				group.Settings = append(group.Settings, SettingItem{
					Label:        kb.HumanLabel(),
					Value:        strings.Join(displayKeys, ", "),
					IsKeybinding: true,
					Action:       kb.Action,
				})
			}
		}
		if len(group.Settings) > 0 {
			groups = append(groups, group)
		}
	}

	// Mark inactive settings
	if !s.UI.ShowSize {
		groups[1].Settings[3].Inactive = true
	}
	if !s.UI.ShowDateModified {
		groups[1].Settings[5].Inactive = true
	}

	return groups
}

func renderGroups(props SettingsProps, groups []SettingGroup) []string {
	var rows []string

	rows = append(rows, "") // Add a line above the first group

	currentIndex := 0
	for i, g := range groups {
		if i > 0 {
			rows = append(rows, "")
		}

		rows = append(rows, props.Style.SettingsHeader.Width(props.Width).Render(g.Title))
		for _, sItem := range g.Settings {

			rows = append(rows, renderSettingRow(props, sItem, currentIndex == props.Cursor))
			currentIndex++
		}
	}
	return rows
}

// renderSettingRow renders a single setting or keybinding row
func renderSettingRow(props SettingsProps, sItem SettingItem, isCursor bool) string {
	val := sItem.Value
	if sItem.Inactive {
		val = props.Style.MutedCol.Render(sItem.Value)
	}

	// For keybindings, color the shortcuts in [] with primary color
	if sItem.IsKeybinding {
		// Use SecondaryCol (usually primary/accent color) for the keys in brackets
		rowStyle := props.Style.SettingsItem
		if isCursor {
			rowStyle = props.Style.SettingsSelectedItem
		}
		val = messages.ColorizeKeysWithStyle(messages.Props{Style: props.Style}, sItem.Value, rowStyle)
	}

	labelWidth := 35
	if props.Width < 60 {
		labelWidth = props.Width - 20
	}
	if labelWidth < 10 {
		labelWidth = 10
	}

	label := ui.Truncate(sItem.Label+":", labelWidth)
	rowContent := fmt.Sprintf("% -*s %s", labelWidth, label, val)

	style := props.Style.SettingsItem
	if isCursor {
		style = props.Style.SettingsSelectedItem
	}

	return style.Width(props.Width).Render(rowContent)
}

// RenderSettingsFooter renders dynamic help text and hints for settings
func RenderSettingsFooter(width int, cursor int, items []SettingHelpItem, styles theme.Stylesheet) string {
	baseFooterStyle := styles.Footer.UnsetPadding().UnsetWidth()

	leftPart := " [↑↓] Navigate | [⏎/Space] Toggle | [r] Reset | [Esc/.] Back"

	helpText := ""
	if cursor >= 0 && cursor < len(items) {
		helpText = items[cursor].HelpText
	}
	rightPart := " " + helpText + " "

	// If width is small, hide shortcuts and show only help text
	if width < 100 {
		return styles.Footer.Width(width).Render(rightPart)
	}

	gap := max(width-lipgloss.Width(leftPart)-lipgloss.Width(rightPart), 0)

	footerContent := messages.ColorizeKeys(messages.Props{Style: styles}, leftPart) + baseFooterStyle.Render(strings.Repeat(" ", gap)) + baseFooterStyle.Render(rightPart)
	return styles.Footer.Width(width).Render(footerContent)
}

// SettingHelpItem represents a setting for help text display
type SettingHelpItem struct {
	HelpText string
}
