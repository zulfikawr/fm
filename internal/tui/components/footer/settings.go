package footer

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderSettingsFooter renders the footer for settings view
func renderSettingsFooter(props Props) string {
	baseFooterStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

	leftPart := " [↑↓] Navigate | [⏎/Space] Toggle | [r] Reset | [Esc/.] Back"
	rightPart := getSettingsHelp(props.SettingsCursor) + " "

	gap := props.Width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart)
	if gap < 0 {
		gap = 0
	}

	footerContent := ColorizeKeys(props, leftPart) + baseFooterStyle.Render(strings.Repeat(" ", gap)) + baseFooterStyle.Render(rightPart)
	return props.Styles.Footer.Width(props.Width).Render(footerContent)
}

// getSettingsHelp returns help text for the current settings cursor position
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
		12: "Change the application color scheme",
	}

	if help, ok := helpTexts[cursor]; ok {
		return help
	}
	return ""
}
