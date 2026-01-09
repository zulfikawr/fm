package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// PrintHelp displays the help screen with keybindings and usage information.
func PrintHelp(s Stylesheet, themeName string) {
	title := s.Header.Render(" FM - Terminal File Manager ")

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	usage := lipgloss.JoinVertical(lipgloss.Left,
		s.DirCol.Render("\nUsage:"),
		fmt.Sprintf("  fm [path]      %s", dim.Render("Open fm in directory")),
		fmt.Sprintf("  fm -h, --help  %s", dim.Render("Show this help screen")),
	)

	keys := lipgloss.JoinVertical(lipgloss.Left,
		s.DirCol.Render("\nKeybindings:"),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("Enter/→/l"), dim.Render("Open directory")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("Backspace/←/h"), dim.Render("Parent directory")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("Space"), dim.Render("Select file")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("s"), dim.Render("Cycle sort mode")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("/"), dim.Render("Search")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("."), dim.Render("Settings & Themes")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("c"), dim.Render("Copy selection")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("v"), dim.Render("Paste clipboard")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("r"), dim.Render("Rename file")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("d"), dim.Render("Delete file")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("Esc"), dim.Render("Unselect / Clear")),
		fmt.Sprintf("  %-20s %s", s.ExecCol.Render("q/Ctrl+C"), dim.Render("Quit")),
	)

	fmt.Println(title)
	fmt.Println(usage)
	fmt.Println(keys)
}
