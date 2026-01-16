package messages

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderInputPrompt renders a labeled text input for the footer area
func RenderInputPrompt(props Props) string {
	input := props.ActiveInput

	// Ensure input styles match footer background
	bg := props.Style.Footer.GetBackground()
	input.TextStyle = props.Style.Footer.UnsetPadding().UnsetWidth()
	input.PromptStyle = props.Style.Footer.UnsetPadding().UnsetWidth()
	input.PlaceholderStyle = props.Style.DimCol.Background(bg)
	// Do not override cursor style here, it's handled by the Input component

	baseStyle := props.Style.Footer.UnsetPadding().UnsetWidth()
	dimStyle := props.Style.DimCol.Inherit(props.Style.Footer).UnsetPadding().UnsetWidth()

	// Update prompt dynamically
	if props.Mode == ModeGoto {
		isRemote := props.AltMode
		if props.RemoteConnected {
			isRemote = !props.AltMode
		}

		label := "Local"
		if isRemote {
			label = "Remote"
		}
		input.Prompt = baseStyle.Render("Go to ") + dimStyle.Render("("+label+")") + baseStyle.Render(": ")
	} else if props.Mode == ModeAuth {
		label := "Password"
		if props.AltMode {
			label = "Path"
		}
		input.Prompt = baseStyle.Render(label + ": ")
	} else if props.Mode == ModeSearching {
		input.Prompt = baseStyle.Render("Filter: ")
	} else if props.Mode == ModeFuzzySearch {
		input.Prompt = baseStyle.Render("Search: ")
	} else if props.Mode == ModeRenaming {
		input.Prompt = baseStyle.Render("Rename: ")
	} else if props.Mode == ModeZip {
		input.Prompt = baseStyle.Render("Zip name: ")
	} else if props.Mode == ModeUnzip {
		input.Prompt = baseStyle.Render("Unzip to: ")
	}

	// Calculate right part (Tab hint) if in Goto, Auth, or FuzzySearch mode
	rightPart := ""
	if props.Mode == ModeGoto || props.Mode == ModeAuth || props.Mode == ModeFuzzySearch {
		dimStyle := props.Style.DimCol.Inherit(props.Style.Footer).UnsetPadding().UnsetWidth()
		keyStyle := props.Style.KeyCol.Inherit(props.Style.Footer).UnsetPadding().UnsetWidth()

		if props.Mode == ModeGoto {
			target := "Remote"
			if props.AltMode {
				target = "Local"
			}
			rightPart = dimStyle.Render("[") + keyStyle.Render("Tab") + dimStyle.Render("] ") + dimStyle.Render(target) + baseStyle.Render(" ")
		} else if props.Mode == ModeAuth {
			target := "Key Path"
			if props.AltMode {
				target = "Password"
			}
			rightPart = dimStyle.Render("[") + keyStyle.Render("Tab") + dimStyle.Render("] ") + dimStyle.Render(target) + baseStyle.Render(" ")
		} else if props.Mode == ModeFuzzySearch {
			rightPart = dimStyle.Render("[") + keyStyle.Render("Tab") + dimStyle.Render("] ") + dimStyle.Render("Collapse") +
				dimStyle.Render(" | [") + keyStyle.Render("Alt+n/m") + dimStyle.Render("] ") + dimStyle.Render("Files") +
				dimStyle.Render(" | [") + keyStyle.Render("Alt+j/k") + dimStyle.Render("] ") + dimStyle.Render("Matches") + baseStyle.Render(" ")
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
	var result strings.Builder
	inBracket := false
	keyStyle := props.Style.KeyCol.Inherit(props.Style.Footer).UnsetPadding().UnsetWidth()
	dimStyle := props.Style.DimCol.Inherit(props.Style.Footer).UnsetPadding().UnsetWidth()
	baseStyle := props.Style.Footer.UnsetPadding().UnsetWidth()

	var current strings.Builder
	for _, r := range str {
		switch r {
		case '[':
			if current.Len() > 0 {
				result.WriteString(baseStyle.Render(current.String()))
				current.Reset()
			}
			inBracket = true
			result.WriteString(dimStyle.Render("["))
		case ']':
			if current.Len() > 0 {
				result.WriteString(keyStyle.Render(current.String()))
				current.Reset()
			}
			inBracket = false
			result.WriteString(dimStyle.Render("]"))
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
