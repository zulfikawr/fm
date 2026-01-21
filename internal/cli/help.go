package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zulfikawr/fm/internal/tui/theme"
)

// PrintHelp displays the help information to the console
func PrintHelp(styles theme.Stylesheet, themeName string) {
	fmt.Println(styles.DirCol.Render("FM - Terminal File Manager"))

	// Define keybindings first to determine max width
	keys := []struct {
		Key  string
		Desc string
	}{
		{"j/down, k/up", "Move cursor"},
		{"l/enter", "Enter directory or open file"},
		{"h/backspace", "Go to parent directory"},
		{"[ / ]", "History Back / Forward"},
		{"Space", "Toggle selection"},
		{"alt+a", "Select all"},
		{"alt+t", "Create new tab"},
		{"alt+n", "Create new file or folder"},
		{"alt+1-9", "Switch to tab 1-9"},
		{"alt+w", "Close current tab"},
		{"alt+l", "Toggle operation logs"},
		{"alt+c", "Toggle clipboard view"},
		{"alt+/", "Fuzzy content search"},
		{"/", "Filter current directory"},
		{"g", "Go to path (local/remote)"},
		{"c", "Copy selected items"},
		{"x", "Cut selected items"},
		{"v", "Paste items from clipboard"},
		{"d", "Delete selected items"},
		{"r", "Rename selected item"},
		{"z", "Zip selected items"},
		{"u", "Unzip selected item"},
		{".", "Toggle settings"},
		{"Esc", "Back / Clear selection"},
		{"ctrl+c", "Quit"},
	}

	// Determine max visible width for alignment
	maxWidth := lipgloss.Width(styles.DirCol.Render("fm") + " " + styles.DimCol.Render("-r user@host[:path]"))

	// Check against keybindings
	for _, k := range keys {
		if w := lipgloss.Width(styles.KeyCol.Render(k.Key)); w > maxWidth {
			maxWidth = w
		}
	}
	maxWidth += 3

	fmt.Println()
	fmt.Println(styles.DirCol.Render("Usage:"))
	usage1Command := styles.GitStaged.Render("fm") + " " + styles.FileCol.Render("[path]")
	fmt.Printf("  %s %s\n", padString(usage1Command, maxWidth), styles.DimCol.Render("Open fm in the specified directory"))
	usage2Command := styles.GitStaged.Render("fm") + " " + styles.FileCol.Render("-r user@host[:path]")
	fmt.Printf("  %s %s\n", padString(usage2Command, maxWidth), styles.DimCol.Render("Open fm on a remote server via SFTP"))
	usage3Command := styles.GitStaged.Render("fm") + " " + styles.GitConflict.Render("search") + " " + styles.FileCol.Render("<query>")
	fmt.Printf("  %s %s\n\n", padString(usage3Command, maxWidth), styles.DimCol.Render("Perform fuzzy search for files and content"))

	fmt.Println(styles.DirCol.Render("Keybindings:"))

	for _, k := range keys {
		// Render the key, calculate its visible width, then pad with spaces
		renderedKey := styles.GitStaged.Render(k.Key)
		visibleWidth := lipgloss.Width(renderedKey)
		padding := strings.Repeat(" ", maxWidth-visibleWidth)
		fmt.Printf("  %s%s %s\n", renderedKey, padding, styles.DimCol.Render(k.Desc))
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
