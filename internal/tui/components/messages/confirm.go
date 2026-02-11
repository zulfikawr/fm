package messages

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/constants"
)

// RenderConfirmationPrompt renders confirmation prompts
func RenderConfirmationPrompt(props Props) string {
	key := fmt.Sprintf("confirm-%s-%d-%s", props.Confirm.ActionType, props.Confirm.ClipboardCount, props.Confirm.ConflictDst)
	if styled, ok := props.Input.PromptCache[key]; ok {
		return props.Style.Footer.Width(props.Width).Render(" " + styled)
	}
	prompt := BuildConfirmationPrompt(props)
	return props.Style.Footer.Width(props.Width).Render(" " + ColorizeKeys(props, prompt))
}

// BuildConfirmationPrompt builds the appropriate confirmation prompt based on action type
func BuildConfirmationPrompt(props Props) string {
	switch props.Confirm.ActionType {
	case constants.ActionDelete:
		return "Delete selected items? [y] Yes | [n] No"
	case constants.ActionPaste:
		if props.Confirm.ClipboardCount == 1 && len(props.Confirm.ClipboardPaths) > 0 {
			// Extract filename from path
			filename := props.Confirm.ClipboardPaths[0]
			if idx := strings.LastIndexAny(filename, "/\\"); idx >= 0 {
				filename = filename[idx+1:]
			}
			return fmt.Sprintf("Paste '%s'? [y] Yes | [n] No", filename)
		}
		return fmt.Sprintf("Paste %d items? [y] Yes | [n] No", props.Confirm.ClipboardCount)
	case constants.ActionResetSettings:
		return "Reset all settings to defaults? [y] Yes | [n] No"
	case constants.ActionRename:
		return fmt.Sprintf("Rename to '%s'? [y] Yes | [n] No", props.Confirm.PendingValue)
	case constants.ActionCreate:
		return fmt.Sprintf("Create '%s'? [y] Yes | [n] No", props.Confirm.PendingValue)
	case constants.ActionZip:
		return fmt.Sprintf("Create zip archive '%s'? [y] Yes | [n] No", props.Confirm.PendingValue)
	case constants.ActionUnzip:
		return fmt.Sprintf("Extract to '%s'? [y] Yes | [n] No", props.Confirm.PendingValue)
	case constants.ActionConflict:
		baseName := extractBaseName(props.Confirm.ConflictDst)
		if props.Confirm.ConflictCount > 1 {
			return fmt.Sprintf("'%s' exists. [y/Y] Overwrite | [n/N] Skip | [r/R] Rename (Upper=All)", baseName)
		}
		return fmt.Sprintf("'%s' exists. [y] Overwrite | [n] Skip | [r] Rename", baseName)
	case constants.ActionTrashRestore:
		baseName := extractBaseName(props.Confirm.ConflictDst)
		return fmt.Sprintf("'%s' exists. [y] Overwrite | [n] Skip | [r] Rename | [c] Cancel", baseName)
	case constants.ActionCancel:
		return "Cancel ongoing operation? [y] Yes | [n] No"
	case constants.ActionUpdate:
		return fmt.Sprintf("A new version of fm (%s) is available. Update? [y] Yes | [n] No", props.Confirm.LatestVersion)
	case constants.ActionGoto:
		return "Go to: [l] Local | [r] Remote"
	case constants.ActionAuth:
		return "Authenticate via: [p] Password | [k] Key"
	case constants.ActionTestIcons:
		return "Do you see these icons correctly?  󰈔  [y] Yes | [n] No"
	default:
		return "Confirm? [y] Yes | [n] No"
	}
}

// RenderHostConfirmPrompt renders host confirmation prompt
func RenderHostConfirmPrompt(props Props) string {
	hostname := ""
	if props.Confirm.HostConfirmReq != nil {
		hostname = props.Confirm.HostConfirmReq.Hostname
	}

	key := "hostconfirm-" + hostname
	if styled, ok := props.Input.PromptCache[key]; ok {
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
