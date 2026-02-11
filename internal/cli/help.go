package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/config"
	"github.com/zulfikawr/fm/internal/tui/components/messages"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// PrintHelp displays the help information to the console
func PrintHelp(styles theme.Stylesheet, themeName string) {
	fmt.Println(styles.DirCol.Render("fm - Terminal File Manager"))

	cfg := config.Load()
	keybinds := cfg.Keybindings

	// Helper to get formatted keys for an action
	getKeys := func(action string) string {
		for i := range keybinds {
			kb := keybinds[i]
			if kb.Action == action {
				displayKeys := make([]string, len(kb.Keys))
				for i, k := range kb.Keys {
					if k == " " {
						displayKeys[i] = "[space]"
					} else {
						displayKeys[i] = "[" + k + "]"
					}
				}
				return strings.Join(displayKeys, ", ")
			}
		}
		return ""
	}

	// Define keybindings sections
	sections := []struct {
		Title string
		Items []struct {
			Key  string
			Desc string
		}
	}{
		{
			Title: "General",
			Items: []struct {
				Key  string
				Desc string
			}{
				{getKeys("quit"), "Quit"},
				{getKeys("help"), "Toggle Help"},
				{getKeys("settings"), "Toggle Settings"},
				{getKeys("analyze"), "Analyze disk usage"},
				{getKeys("clipboard_view"), "Toggle Clipboard"},
				{getKeys("logs_view"), "Toggle Logs"},
			},
		},
		{
			Title: "Navigation",
			Items: []struct {
				Key  string
				Desc string
			}{
				{getKeys("open"), "Open / Enter"},
				{getKeys("go_parent"), "Go to Parent"},
				{getKeys("move_down") + ", " + getKeys("move_up"), "Move cursor"},
				{"[Shift+j/k]", "Range selection"},
				{getKeys("go_to_path"), "Go to Path"},
				{getKeys("history_back") + ", " + getKeys("history_forward"), "History Back / Forward"},
				{getKeys("cycle_sort"), "Cycle Sort"},
			},
		},
		{
			Title: "File Operations",
			Items: []struct {
				Key  string
				Desc string
			}{
				{getKeys("copy"), "Copy"},
				{getKeys("cut"), "Cut"},
				{getKeys("paste"), "Paste"},
				{getKeys("rename"), "Rename"},
				{getKeys("delete"), "Delete"},
				{getKeys("create"), "Create New File/Folder"},
				{getKeys("zip"), "Zip"},
				{getKeys("unzip"), "Unzip"},
			},
		},
		{
			Title: "Selection",
			Items: []struct {
				Key  string
				Desc string
			}{
				{getKeys("toggle_selection"), "Toggle Selection"},
				{getKeys("select_all"), "Select All"},
				{getKeys("clear_selection"), "Clear Selection"},
			},
		},
		{
			Title: "Search & Filter",
			Items: []struct {
				Key  string
				Desc string
			}{
				{getKeys("filter"), "Filter Directory"},
				{"[tab]", "Autocomplete Name/Path"},
				{getKeys("fuzzy_search"), "Fuzzy Search"},
				{getKeys("toggle_regex_search"), "Toggle Regex Search"},
			},
		},
		{
			Title: "Tabs",
			Items: []struct {
				Key  string
				Desc string
			}{
				{getKeys("new_tab"), "New Tab"},
				{getKeys("close_tab"), "Close Tab"},
				{"[alt+1-9]", "Switch to Tab 1-9"},
			},
		},
	}

	// Determine max visible width for alignment
	usageWidth := lipgloss.Width(styles.GitStaged.Render("fm")+" "+styles.GitConflict.Render("search")+" "+styles.DimCol.Render("[--regex]")+" "+styles.FileCol.Render("<query>")) + 3

	keyWidth := 0
	for i := range sections {
		s := sections[i]
		for j := range s.Items {
			k := s.Items[j]
			if w := lipgloss.Width(styles.DimCol.Render(k.Desc)); w > keyWidth {
				keyWidth = w
			}
		}
	}
	keyWidth += 3

	fmt.Println()
	fmt.Println(styles.DirCol.Render("Usage:"))
	usage1Command := styles.GitStaged.Render("fm") + " " + styles.FileCol.Render("[path]")
	fmt.Printf("  %s %s\n", padString(usage1Command, usageWidth), styles.DimCol.Render("Open fm in the specified directory"))
	usage2Command := styles.GitStaged.Render("fm") + " " + styles.FileCol.Render("-r user@host[:path]")
	fmt.Printf("  %s %s\n", padString(usage2Command, usageWidth), styles.DimCol.Render("Open fm on a remote server via SFTP"))
	usage3Command := styles.GitStaged.Render("fm") + " " + styles.GitConflict.Render("search") + " " + styles.DimCol.Render("[--regex]") + " " + styles.FileCol.Render("<query>")
	fmt.Printf("  %s %s\n", padString(usage3Command, usageWidth), styles.DimCol.Render("Perform fuzzy or regex search for files and content"))
	usage4Command := styles.GitStaged.Render("fm") + " " + styles.GitConflict.Render("info") + " " + styles.FileCol.Render("[path]")
	fmt.Printf("  %s %s\n", padString(usage4Command, usageWidth), styles.DimCol.Render("Show file/directory information"))
	usage5Command := styles.GitStaged.Render("fm") + " " + styles.GitConflict.Render("analyze") + " " + styles.FileCol.Render("[path]")
	fmt.Printf("  %s %s\n", padString(usage5Command, usageWidth), styles.DimCol.Render("Analyze disk usage of a directory"))
	usage6Command := styles.GitStaged.Render("fm") + " " + styles.GitConflict.Render("config") + " " + styles.DimCol.Render("[--reset | init]")
	fmt.Printf("  %s %s\n\n", padString(usage6Command, usageWidth), styles.DimCol.Render("Manage configuration (view, reset, or interactive init)"))

	for i := range sections {
		s := sections[i]
		fmt.Println(styles.DirCol.Render(s.Title + ":"))
		for j := range s.Items {
			k := s.Items[j]
			// Render the description first, then pad
			renderedDesc := styles.DimCol.Render(k.Desc)
			visibleWidth := lipgloss.Width(renderedDesc)
			padding := strings.Repeat(" ", keyWidth-visibleWidth)

			// Colorize keys in brackets using the same logic as the TUI
			// Use an empty base style to avoid background inheritance in the CLI
			props := messages.Props{Style: styles}
			coloredKey := messages.ColorizeKeysWithStyle(props, k.Key, lipgloss.NewStyle())

			fmt.Printf("  %s%s %s\n", renderedDesc, padding, coloredKey)
		}
		fmt.Println()
	}
}

// padString takes a string that might contain ANSI codes, and pads it to the targetWidth based on its visible width.
func padString(s string, targetWidth int) string {
	visibleWidth := lipgloss.Width(s)
	if visibleWidth >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-visibleWidth)
}
