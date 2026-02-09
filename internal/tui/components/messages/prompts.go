package messages

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderInputPrompt renders a labeled text input for the footer area
func RenderInputPrompt(props Props) string {
	input := props.Input.Active

	// Ensure input styles match footer background
	bg := props.Style.Footer.GetBackground()
	input.TextStyle = props.Style.Footer.UnsetPadding().UnsetWidth()
	input.PromptStyle = props.Style.Footer.UnsetPadding().UnsetWidth()
	input.PlaceholderStyle = props.Style.MutedCol.Background(bg)
	// Do not override cursor style here, it's handled by the Input component

	baseStyle := props.Style.Footer.UnsetPadding().UnsetWidth()
	mutedStyle := props.Style.MutedCol.Inherit(props.Style.Footer).UnsetPadding().UnsetWidth()

	// Update prompt dynamically
	switch props.Mode {
	case ModeGoto:
		label := "Local"
		if props.Input.AltMode {
			label = "Remote"
		}
		input.Prompt = baseStyle.Render("Go to ") + mutedStyle.Render("("+label+")") + baseStyle.Render(": ")
	case ModeAuth:
		label := "Password"
		if props.Input.AltMode {
			label = "PEM Path"
		}
		input.Prompt = baseStyle.Render(label + ": ")
	case ModeSearching:
		input.Prompt = baseStyle.Render("Filter: ")
	case ModeFuzzySearch:
		input.Prompt = baseStyle.Render("Search: ")
	case ModeRenaming:
		input.Prompt = baseStyle.Render("Rename: ")
	case ModeZip:
		input.Prompt = baseStyle.Render("Zip name: ")
	case ModeUnzip:
		input.Prompt = baseStyle.Render("Unzip to: ")
	case ModeCreate:
		label := "File"
		if props.Input.AltMode {
			label = "Folder"
		}
		input.Prompt = baseStyle.Render("Create ") + mutedStyle.Render("("+label+")") + baseStyle.Render(": ")
	case ModeConflictRename:
		input.Prompt = baseStyle.Render("New name: ")
	}

	// Calculate right part (Tab hint) if in FuzzySearch, or Create mode
	rightPart := ""
	if props.Mode == ModeFuzzySearch || props.Mode == ModeCreate {
		mutedStyle := props.Style.MutedCol.Inherit(props.Style.Footer).UnsetPadding().UnsetWidth()
		accentStyle := props.Style.AccentCol.Inherit(props.Style.Footer).UnsetPadding().UnsetWidth()

		switch props.Mode {
		case ModeCreate:
			target := "Folder"
			if props.Input.AltMode {
				target = "File"
			}
			rightPart = mutedStyle.Render("[") + accentStyle.Render("Tab") + mutedStyle.Render("] ") + mutedStyle.Render(target) + baseStyle.Render(" ")
		case ModeFuzzySearch:
			rightPart = mutedStyle.Render("[") + accentStyle.Render("Tab") + mutedStyle.Render("] ") + mutedStyle.Render("Collapse") +
				mutedStyle.Render(" | [") + accentStyle.Render("Alt+n/m") + mutedStyle.Render("] ") + mutedStyle.Render("Files") +
				mutedStyle.Render(" | [") + accentStyle.Render("Alt+j/k") + mutedStyle.Render("] ") + mutedStyle.Render("Matches") + baseStyle.Render(" ")
		}
	}

	// Adjust input width to fit within props.Width
	rightWidth := lipgloss.Width(rightPart)
	promptWidth := lipgloss.Width(input.Prompt)
	availableInputWidth := props.Width - 1 - promptWidth - rightWidth - 1
	if availableInputWidth < 5 {
		availableInputWidth = 5
	}
	if input.Width > availableInputWidth {
		input.Width = availableInputWidth
	}

	// Calculate left part (input)
	leftPart := baseStyle.Render(" ") + input.View()

	// Calculate gap
	leftWidth := lipgloss.Width(leftPart)
	gapWidth := props.Width - leftWidth - rightWidth
	if gapWidth < 0 {
		gapWidth = 0
	}

	gap := baseStyle.Render(strings.Repeat(" ", gapWidth))

	content := leftPart + gap + rightPart
	return props.Style.Footer.Width(props.Width).Render(content)
}

// ColorizeKeys colorizes key indicators in brackets
func ColorizeKeys(props Props, str string) string {
	return ColorizeKeysWithStyle(props, str, props.Style.Footer)
}

// ColorizeKeysWithStyle colorizes key indicators in brackets with a specific base style
func ColorizeKeysWithStyle(props Props, str string, base lipgloss.Style) string {
	var result strings.Builder
	inBracket := false

	// Extract the background from the base style to ensure consistency
	bg := base.GetBackground()

	keyStyle := props.Style.KeyCol.Inherit(base).UnsetPadding().UnsetWidth().Background(bg)
	dimStyle := props.Style.DimCol.Inherit(base).UnsetPadding().UnsetWidth().Background(bg)
	baseStyle := base.UnsetPadding().UnsetWidth()

	var current strings.Builder
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch r {
		case '[':
			if !inBracket {
				if current.Len() > 0 {
					result.WriteString(baseStyle.Render(current.String()))
					current.Reset()
				}
				inBracket = true
				result.WriteString(dimStyle.Render("["))
			} else {
				// Nested bracket or bracket as content
				current.WriteRune(r)
			}
		case ']':
			if inBracket {
				// Heuristic: if there's another ']' immediately following this one,
				// then this one is likely part of the key content (e.g., in "[]]")
				if i+1 < len(runes) && runes[i+1] == ']' {
					current.WriteRune(r)
				} else {
					// This is the closing bracket
					result.WriteString(keyStyle.Render(current.String()))
					current.Reset()
					inBracket = false
					result.WriteString(dimStyle.Render("]"))
				}
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		if inBracket {
			result.WriteString(keyStyle.Render(current.String()))
		} else {
			result.WriteString(baseStyle.Render(current.String()))
		}
	}
	return result.String()
}
