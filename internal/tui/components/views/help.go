package views

import (
	"strings"

	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/tui/components/messages"
	"github.com/zulfikawr/fm/internal/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// HelpProps contains data for rendering the help view
type HelpProps struct {
	Width       int
	Height      int
	Cursor      int
	Offset      int
	Style       theme.Stylesheet
	Keybindings []config.Keybinding
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

	groups := buildHelpGroups(props.Keybindings)
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

func buildHelpGroups(keybinds []config.Keybinding) []HelpSection {
	// Helper to get formatted keys for an action
	getKeys := func(action string) string {
		for _, kb := range keybinds {
			if kb.Action == action {
				displayKeys := make([]string, len(kb.Keys))
				for i, k := range kb.Keys {
					if k == " " {
						displayKeys[i] = "[space]"
					} else {
						displayKeys[i] = "[" + k + "]"
					}
				}
				return strings.Join(displayKeys, "/")
			}
		}
		return ""
	}

	return []HelpSection{
		{
			Title: "General",
			Items: []HelpItem{
				{Key: getKeys("quit"), Desc: "Quit fm"},
				{Key: getKeys("help"), Desc: "Toggle this help screen"},
				{Key: getKeys("settings"), Desc: "Toggle settings menu"},
				{Key: getKeys("analyze"), Desc: "Analyze Disk Usage"},
				{Key: getKeys("clipboard_view"), Desc: "View current clipboard contents"},
				{Key: getKeys("logs_view"), Desc: "View operation logs"},
			},
		},
		{
			Title: "Navigation",
			Items: []HelpItem{
				{Key: getKeys("open"), Desc: "Open directory / Open file in editor"},
				{Key: getKeys("go_parent"), Desc: "Navigate to parent directory"},
				{Key: getKeys("move_down") + ", " + getKeys("move_up"), Desc: "Move selection down / up"},
				{Key: "[Shift+j/↓], [Shift+k/↑]", Desc: "Range selection"},
				{Key: getKeys("go_to_path"), Desc: "Go to path (choose Local [l] or Remote [r])"},
				{Key: getKeys("history_back"), Desc: "History Back"},
				{Key: getKeys("history_forward"), Desc: "History Forward"},
				{Key: getKeys("cycle_sort"), Desc: "Cycle through sort modes"},
			},
		},
		{
			Title: "File Operations",
			Items: []HelpItem{
				{Key: getKeys("copy"), Desc: "Copy selected items to clipboard"},
				{Key: getKeys("cut"), Desc: "Cut selected items to clipboard"},
				{Key: getKeys("paste"), Desc: "Paste items from clipboard"},
				{Key: getKeys("rename"), Desc: "Rename highlighted item"},
				{Key: getKeys("delete"), Desc: "Delete selected items"},
				{Key: getKeys("create"), Desc: "Create new item (File/Folder)"},
				{Key: getKeys("zip"), Desc: "Create Zip archive"},
				{Key: getKeys("unzip"), Desc: "Unzip highlighted archive"},
			},
		},
		{
			Title: "Selection",
			Items: []HelpItem{
				{Key: getKeys("toggle_selection"), Desc: "Toggle selection"},
				{Key: "[Shift+Left Click]", Desc: "Toggle selection / Range select"},
				{Key: getKeys("select_all"), Desc: "Select all items in directory"},
				{Key: getKeys("clear_selection"), Desc: "Clear all selections"},
			},
		},
		{
			Title: "Search & Filter",
			Items: []HelpItem{
				{Key: getKeys("filter"), Desc: "Enter filter mode (↑/↓ to navigate)"},
				{Key: "[tab]", Desc: "Autocomplete current name or path"},
				{Key: getKeys("fuzzy_search"), Desc: "Fuzzy Content Search (Find in Files)"},
				{Key: getKeys("toggle_regex_search"), Desc: "Toggle Regex Search mode"},
				{Key: "[Alt+M/Alt+N]", Desc: "Jump between files in search results"},
			},
		},
		{
			Title: "Tabs",
			Items: []HelpItem{
				{Key: getKeys("new_tab"), Desc: "Open a new tab (up to 9)"},
				{Key: getKeys("close_tab"), Desc: "Close current tab"},
				{Key: "[Alt+1-9]", Desc: "Switch to corresponding tab"},
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

	keyStyle := props.Style.SecondaryCol.UnsetPadding().UnsetWidth()
	descStyle := props.Style.MutedCol.UnsetPadding().UnsetWidth()
	rowStyle := props.Style.Item

	if isCursor {
		rowStyle = props.Style.SelectedItem.UnsetPadding().UnsetWidth()
		keyStyle = keyStyle.Inherit(rowStyle)
		descStyle = descStyle.Inherit(rowStyle)
	}

	// To ensure full-width highlight, we construct the row as a single string
	// with explicit padding and then render it with rowStyle.

	// Key part with fixed width and padding
	keyContent := keyStyle.Render(item.Key)

	leftPart := rowStyle.PaddingLeft(3).Width(keyWidth).Render(keyContent)

	// Desc part with remaining width
	descWidth := max(props.Width-keyWidth, 10)
	descContent := descStyle.Render(item.Desc)

	rightPart := rowStyle.Width(descWidth).Render(descContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPart, rightPart)
}

// RenderHelpFooter renders hints for the help view
func RenderHelpFooter(width int, styles theme.Stylesheet) string {
	hint := "[Esc/?] Back"
	return styles.Footer.Width(width).Render(" " + messages.ColorizeKeys(messages.Props{Style: styles}, hint))
}
