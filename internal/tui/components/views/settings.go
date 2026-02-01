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
	Label    string
	Value    string
	Inactive bool
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
				{Label: "Show Hidden Files", Value: ui.Toggle(s.ShowHidden, styles)},
				{Label: "Case-Sensitive Search", Value: ui.Toggle(s.CaseSensitive, styles)},
				{Label: "Confirm Operations", Value: ui.Toggle(s.ConfirmOperations, styles)},
				{Label: "Wrap Navigation", Value: ui.Toggle(s.WrapNavigation, styles)},
				{Label: "Preferred Editor", Value: ui.Picker(constants.Editors[s.EditorIndex], styles)},
				{Label: "Use Trash (Move to Trash)", Value: ui.Toggle(s.UseTrash, styles)},
			},
		},
		{
			Title: "Display Options",
			Settings: []SettingItem{
				{Label: "Show Column Headers", Value: ui.Toggle(s.ShowHeader, styles)},
				{Label: "Enable Git Status", Value: ui.Toggle(s.EnableGit, styles)},
				{Label: "Show File Size", Value: ui.Toggle(s.ShowSize, styles)},
				{Label: "Size Format", Value: ui.Picker(format.SizeFormats[s.SizeFormatIndex], styles)},
				{Label: "Show Date Modified", Value: ui.Toggle(s.ShowDateModified, styles)},
				{Label: "Date Format", Value: ui.Picker(format.DateFormats[s.DateFormatIndex].Name, styles)},
				{Label: "Enable Mouse Support", Value: ui.Toggle(s.EnableMouse, styles)},
			},
		},
		{
			Title: "Appearance",
			Settings: []SettingItem{
				{Label: "Enable Nerd Font Icons", Value: ui.Toggle(s.EnableIcons, styles)},
				{Label: "Theme", Value: ui.Picker(theme.Themes[s.ThemeIndex].Name, styles)},
			},
		},
	}
	// Mark inactive settings
	if !s.ShowSize {
		groups[1].Settings[3].Inactive = true
	}
	if !s.ShowDateModified {
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

func renderSettingRow(props SettingsProps, sItem SettingItem, isCursor bool) string {
	val := sItem.Value
	if sItem.Inactive {
		val = props.Style.MutedCol.Render(sItem.Value)
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
func RenderSettingsFooter(width int, cursor int, styles theme.Stylesheet) string {
	baseFooterStyle := styles.Footer.UnsetPadding().UnsetWidth()

	leftPart := " [↑↓] Navigate | [⏎/Space] Toggle | [r] Reset | [Esc/.] Back"
	rightPart := " " + getSettingsHelp(cursor) + " "

	// If width is small, hide shortcuts and show only help text
	if width < 100 {
		return styles.Footer.Width(width).Render(rightPart)
	}

	gap := max(width-lipgloss.Width(leftPart)-lipgloss.Width(rightPart), 0)

	footerContent := messages.ColorizeKeys(messages.Props{Style: styles}, leftPart) + baseFooterStyle.Render(strings.Repeat(" ", gap)) + baseFooterStyle.Render(rightPart)
	return styles.Footer.Width(width).Render(footerContent)
}

func getSettingsHelp(cursor int) string {
	helpTexts := map[int]string{
		0:  "Show/hide files starting with '.'",
		1:  "Search respects capitalization",
		2:  "Ask before destructive actions",
		3:  "Cursor loops at list boundaries",
		4:  "Choose default editor for opening files",
		5:  "Move deleted items to trash (off = permanent delete)",
		6:  "Show/hide list column headers",
		7:  "Enable git status markers",
		8:  "Show file size in list",
		9:  "Change the file size display unit",
		10: "Show last modification time",
		11: "Change the date and time format",
		12: "Allow mouse interaction (clicks, scroll)",
		13: "Toggle Nerd Font icons (requires download)",
		14: "Change the application color scheme",
	}

	if help, ok := helpTexts[cursor]; ok {
		return help
	}
	return ""
}
