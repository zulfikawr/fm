package views

import (
	"strings"

	"github.com/zulfikawr/fm/internal/tui/components/messages"
	"github.com/zulfikawr/fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// HelpProps contains data for rendering the help view
type HelpProps struct {
	Width  int
	Height int
	Cursor int
	Offset int
	Style  theme.Stylesheet
}

// HelpSection represents a categorized group of keybindings
type HelpSection struct {
	Title string
	Items []HelpItem
}

// HelpItem represents a single keybinding entry
type HelpItem struct {
	Key  string
	Desc string
}

// RenderHelp renders the help menu view
func RenderHelp(props HelpProps) string {
	if props.Height <= 0 {
		return ""
	}

	groups := buildHelpGroups()
	rows := renderHelpRows(props, groups)

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

func buildHelpGroups() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{Key: "Enter/→/l", Desc: "Open directory / Open file in editor"},
				{Key: "Backspace/←/h", Desc: "Navigate to parent directory"},
				{Key: "j/↓, k/↑", Desc: "Move selection down / up"},
				{Key: "Shift+j/↓, Shift+k/↑", Desc: "Range selection"},
				{Key: "g", Desc: "Go to path (choose Local [l] or Remote [r])"},
				{Key: "[", Desc: "History Back"},
				{Key: "]", Desc: "History Forward"},
			},
		},
		{
			Title: "Selection & Bulk Actions",
			Items: []HelpItem{
				{Key: "Space", Desc: "Toggle selection"},
				{Key: "Shift+Left Click", Desc: "Toggle selection / Range select"},
				{Key: "Alt+A", Desc: "Select all items in directory"},
				{Key: "Esc", Desc: "Clear all selections"},
			},
		},
		{
			Title: "Tabs",
			Items: []HelpItem{
				{Key: "Alt+T", Desc: "Open a new tab (up to 9)"},
				{Key: "Alt+W", Desc: "Close current tab"},
				{Key: "Alt+1-9", Desc: "Switch to corresponding tab"},
			},
		},
		{
			Title: "File Operations",
			Items: []HelpItem{
				{Key: "c", Desc: "Copy selected items to clipboard"},
				{Key: "x", Desc: "Cut selected items to clipboard"},
				{Key: "v", Desc: "Paste items from clipboard"},
				{Key: "r", Desc: "Rename highlighted item"},
				{Key: "d", Desc: "Delete selected items"},
				{Key: "Alt+N", Desc: "Create new item (File/Folder)"},
				{Key: "z", Desc: "Create Zip archive"},
				{Key: "u", Desc: "Unzip highlighted archive"},
			},
		},
		{
			Title: "Search, Filtering & Inputs",
			Items: []HelpItem{
				{Key: "/", Desc: "Enter filter mode (↑/↓ to navigate)"},
				{Key: "Tab", Desc: "Autocomplete current name or path"},
				{Key: "Alt+/", Desc: "Fuzzy Content Search (Find in Files)"},
				{Key: "Alt+M/Alt+N", Desc: "Jump between files in search results"},
				{Key: "Esc", Desc: "Exit search/filter mode"},
			},
		},
		{
			Title: "Miscellaneous", Items: []HelpItem{
				{Key: "Alt+C", Desc: "View current clipboard contents"},
				{Key: "Alt+L", Desc: "View operation logs"},
				{Key: "s", Desc: "Cycle through sort modes"},
				{Key: ".", Desc: "Toggle settings menu"},
				{Key: "?", Desc: "Toggle this help screen"},
				{Key: "Ctrl+C", Desc: "Quit fm"},
			},
		},
	}
}

func renderHelpRows(props HelpProps, groups []HelpSection) []string {
	var rows []string
	rows = append(rows, "") // Top margin

	currentIndex := 0
	for _, s := range groups {
		rows = append(rows, props.Style.SettingsHeader.Width(props.Width).Render(s.Title))
		for _, item := range s.Items {
			rows = append(rows, renderHelpRow(props, item, currentIndex == props.Cursor))
			currentIndex++
		}
		rows = append(rows, "") // Spacer
	}
	return rows
}

func renderHelpRow(props HelpProps, item HelpItem, isCursor bool) string {
	keyWidth := 25
	if props.Width < 60 {
		keyWidth = props.Width / 3
	}

	keyStyle := props.Style.GitStaged.UnsetPadding().UnsetWidth()
	descStyle := props.Style.DimCol.UnsetPadding().UnsetWidth()

	key := keyStyle.Render(item.Key)
	desc := descStyle.Render(item.Desc)

	rowContent := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(keyWidth).PaddingLeft(3).Render(key),
		lipgloss.NewStyle().Width(props.Width-keyWidth-3).Render(desc),
	)

	return props.Style.Item.Width(props.Width).Render(rowContent)
}

// RenderHelpFooter renders hints for the help view
func RenderHelpFooter(width int, styles theme.Stylesheet) string {
	hint := "[Esc/?] Back"
	return styles.Footer.Width(width).Render(" " + messages.ColorizeKeys(messages.Props{Style: styles}, hint))
}
