package footer

import (
	"fmt"
	"strings"

	"fm/internal/constants"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// renderPrompts renders the appropriate prompt (input or confirmation) based on mode
func renderPrompts(props Props) string {
	switch props.Mode {
	case ModeSearching, ModeRenaming, ModeGoto, ModeAuth:
		return renderInputPrompt(props, props.ActiveInput)
	case ModeConfirming:
		return renderConfirmationPrompt(props)
	case ModeHostConfirm:
		return renderHostConfirmPrompt(props)
	default:
		return ""
	}
}

// renderInputPrompt renders a footer with an input field
func renderInputPrompt(props Props, input textinput.Model) string {
	// Ensure input styles match footer background
	bg := props.Styles.Footer.GetBackground()
	input.TextStyle = props.Styles.Footer.UnsetPadding().UnsetWidth()
	input.PromptStyle = props.Styles.Footer.UnsetPadding().UnsetWidth()
	input.PlaceholderStyle = props.Styles.DimCol.Background(bg)
	input.CursorStyle = props.Styles.Footer.UnsetPadding().UnsetWidth()

	baseStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()
	dimStyle := props.Styles.DimCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()

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
	}

	// Calculate right part (Tab hint) if in Goto or Auth mode
	rightPart := ""
	if props.Mode == ModeGoto || props.Mode == ModeAuth {
		dimStyle := props.Styles.DimCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()
		keyStyle := props.Styles.KeyCol.Inherit(props.Styles.Footer).UnsetPadding().UnsetWidth()

		target := ""
		if props.Mode == ModeGoto {
			target = "Remote"
			if props.AltMode {
				target = "Local"
			}
		} else if props.Mode == ModeAuth {
			target = "Key Path"
			if props.AltMode {
				target = "Password"
			}
		}

		rightPart = dimStyle.Render("[") + keyStyle.Render("Tab") + dimStyle.Render("] ") + dimStyle.Render(target) + baseStyle.Render(" ")
	}

	// Adjust input width to fit within props.Width
	rightWidth := lipgloss.Width(rightPart)
	promptWidth := lipgloss.Width(input.Prompt)
	// Available width for the input text itself: total - leading space - prompt - right part - margin
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
	return props.Styles.Footer.Width(props.Width).Render(content)
}

// renderConfirmationPrompt renders confirmation prompts
func renderConfirmationPrompt(props Props) string {
	key := fmt.Sprintf("confirm-%s-%d-%s", props.ActionType, props.ClipboardCount, props.ConflictDst)
	if styled, ok := props.PromptCache[key]; ok {
		return props.Styles.Footer.Width(props.Width).Render(" " + styled)
	}
	prompt := BuildConfirmationPrompt(props)
	return props.Styles.Footer.Width(props.Width).Render(" " + ColorizeKeys(props, prompt))
}

// BuildConfirmationPrompt builds the appropriate confirmation prompt based on action type
func BuildConfirmationPrompt(props Props) string {
	switch props.ActionType {
	case constants.ActionDelete:
		return "Delete selected items? (y/n)"
	case constants.ActionPaste:
		return fmt.Sprintf("Paste %d items? (y/n)", props.ClipboardCount)
	case constants.ActionResetSettings:
		return "Reset all settings to defaults? (y/n)"
	case constants.ActionConflict:
		baseName := extractBaseName(props.ConflictDst)
		return fmt.Sprintf("'%s' exists. [y] Overwrite | [n] Skip | [r] Rename", baseName)
	case constants.ActionCancel:
		return "Cancel ongoing operation? (y/n)"
	default:
		return "Confirm? (y/n)"
	}
}

// renderHostConfirmPrompt renders host confirmation prompt
func renderHostConfirmPrompt(props Props) string {
	hostname := ""
	if props.HostConfirmReq != nil {
		hostname = props.HostConfirmReq.Hostname
	}

	key := "hostconfirm-" + hostname
	if styled, ok := props.PromptCache[key]; ok {
		return props.Styles.Footer.Width(props.Width).Render(" " + styled)
	}

	prompt := fmt.Sprintf("Add host '%s' to known_hosts? (y/n)", hostname)
	return props.Styles.Footer.Width(props.Width).Render(" " + ColorizeKeys(props, prompt))
}

// extractBaseName extracts the base name from a path
func extractBaseName(path string) string {
	baseName := path
	if idx := strings.LastIndexAny(path, "/\\"); idx != -1 {
		baseName = path[idx+1:]
	}
	return baseName
}

// ColorizeKeys colorizes key indicators in brackets
func ColorizeKeys(props Props, str string) string {
	var result strings.Builder
	inBracket := false
	keyStyle := props.Styles.KeyCol.Inherit(props.Styles.Footer)
	baseStyle := props.Styles.Footer.UnsetPadding().UnsetWidth()

	var current strings.Builder
	for _, r := range str {
		switch r {
		case '[':
			if current.Len() > 0 {
				result.WriteString(baseStyle.Render(current.String()))
				current.Reset()
			}
			inBracket = true
			result.WriteString(keyStyle.Render("["))
		case ']':
			if current.Len() > 0 {
				result.WriteString(keyStyle.Render(current.String()))
				current.Reset()
			}
			inBracket = false
			result.WriteString(keyStyle.Render("]"))
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
