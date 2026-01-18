package messages

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
)

// RenderConfirmationPrompt renders confirmation prompts
func RenderConfirmationPrompt(props Props) string {
	key := fmt.Sprintf("confirm-%s-%d-%s", props.ActionType, props.ClipboardCount, props.ConflictDst)
	if styled, ok := props.PromptCache[key]; ok {
		return props.Style.Footer.Width(props.Width).Render(" " + styled)
	}
	prompt := BuildConfirmationPrompt(props)
	return props.Style.Footer.Width(props.Width).Render(" " + ColorizeKeys(props, prompt))
}

// BuildConfirmationPrompt builds the appropriate confirmation prompt based on action type
func BuildConfirmationPrompt(props Props) string {
	switch props.ActionType {
	case constants.ActionDelete:
		return "Delete selected items? [y] Yes | [n] No"
	case constants.ActionPaste:
		return fmt.Sprintf("Paste %d items? [y] Yes | [n] No", props.ClipboardCount)
	case constants.ActionResetSettings:
		return "Reset all settings to defaults? [y] Yes | [n] No"
	case constants.ActionConflict:
		baseName := extractBaseName(props.ConflictDst)
		if props.ConflictPendingCount > 1 {
			return fmt.Sprintf("'%s' exists. [y/Y] Overwrite | [n/N] Skip | [r/R] Rename (Upper=All)", baseName)
		}
		return fmt.Sprintf("'%s' exists. [y] Overwrite | [n] Skip | [r] Rename", baseName)
	case constants.ActionCancel:
		return "Cancel ongoing operation? [y] Yes | [n] No"
	default:
		return "Confirm? [y] Yes | [n] No"
	}
}

// RenderHostConfirmPrompt renders host confirmation prompt
func RenderHostConfirmPrompt(props Props) string {
	hostname := ""
	if props.HostConfirmReq != nil {
		hostname = props.HostConfirmReq.Hostname
	}

	key := "hostconfirm-" + hostname
	if styled, ok := props.PromptCache[key]; ok {
		return props.Style.Footer.Width(props.Width).Render(" " + styled)
	}

	prompt := fmt.Sprintf("Add host '%s' to known_hosts? [y] Yes | [n] No", hostname)
	return props.Style.Footer.Width(props.Width).Render(" " + ColorizeKeys(props, prompt))
}

func extractBaseName(path string) string {
	baseName := path
	if idx := strings.LastIndexAny(path, "/\\"); idx != -1 {
		baseName = path[idx+1:]
	}
	return baseName
}
