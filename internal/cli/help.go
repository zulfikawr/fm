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
				{"[1-9]", "Switch to Tab 1-9"},
			},
		},
	}

	// Determine max visible width for alignment
	usageLines := []struct {
		command string
		desc    string
	}{
		{"[path]", "Open fm in the specified directory"},
		{"-r | --remote user@host[:path]", "Open fm on a remote server via SFTP"},
		{"-s | --search <query> [-e | --regex]", "Perform fuzzy or regex search for files and content"},
		{"-i | --info [path] [--json] [--tree] [--depth N]", "Show file/directory information"},
		{"-a | --analyze [path]", "Analyze disk usage of a directory"},
		{"-c | --config [--reset | --init]", "Manage configuration (view, reset, or interactive init)"},
	}

	// Helper to colorize command parts
	colorizeCommand := func(cmd string) string {
		result := ""
		inBracket := false
		inAngle := false
		word := ""

		for _, ch := range cmd {
			switch ch {
			case '[':
				if word != "" {
					result += styles.DimCol.Render(word)
				}
				inBracket = true
				word = "["
			case ']':
				word += "]"
				result += styles.FileCol.Render(word)
				word = ""
				inBracket = false
			case '<':
				if word != "" {
					result += styles.DimCol.Render(word)
				}
				inAngle = true
				word = "<"
			case '>':
				word += ">"
				result += styles.AccentCol.Render(word)
				word = ""
				inAngle = false
			case ' ', '|':
				if word != "" {
					if inBracket || inAngle {
						word += string(ch)
					} else if word == "-r" || word == "-s" || word == "-i" || word == "-a" || word == "-c" || word == "-e" {
						result += styles.GitStaged.Render(word)
						word = ""
						result += styles.DimCol.Render(string(ch))
					} else if word == "--remote" || word == "--search" || word == "--info" || word == "--analyze" || word == "--config" || word == "--regex" || word == "--json" || word == "--tree" || word == "--depth" || word == "--reset" || word == "--init" {
						result += styles.GitStaged.Render(word)
						word = ""
						result += styles.DimCol.Render(string(ch))
					} else {
						result += styles.DimCol.Render(word)
						word = ""
						result += styles.DimCol.Render(string(ch))
					}
				} else {
					result += styles.DimCol.Render(string(ch))
				}
			default:
				word += string(ch)
			}
		}
		if word != "" {
			result += styles.DimCol.Render(word)
		}
		return result
	}

	// Calculate max command width for alignment
	maxCmdWidth := 0
	for i := range usageLines {
		cmdRendered := styles.PrimaryCol.Render("fm") + " " + colorizeCommand(usageLines[i].command)
		w := lipgloss.Width(cmdRendered)
		if w > maxCmdWidth {
			maxCmdWidth = w
		}
	}
	maxCmdWidth += 3 // Add padding

	keyWidth := 0
	for i := range sections {
		s := sections[i]
		for j := range s.Items {
			k := s.Items[j]
			if w := lipgloss.Width(styles.FileCol.Render(k.Desc)); w > keyWidth {
				keyWidth = w
			}
		}
	}
	keyWidth += 3

	fmt.Println()
	fmt.Println(styles.DirCol.Render("Usage:"))
	for i := range usageLines {
		line := usageLines[i]
		cmdRendered := styles.PrimaryCol.Render("fm") + " " + colorizeCommand(line.command)
		fmt.Printf("  %s %s\n", padString(cmdRendered, maxCmdWidth), styles.DimCol.Render(line.desc))
	}
	fmt.Println()

	for i := range sections {
		s := sections[i]
		fmt.Println(styles.DirCol.Render(s.Title + ":"))
		for j := range s.Items {
			k := s.Items[j]
			// Render the description first, then pad
			renderedDesc := styles.FileCol.Render(k.Desc)
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
