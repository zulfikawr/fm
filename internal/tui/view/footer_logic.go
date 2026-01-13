package view

import (
	"fmt"
	"strings"

	"fm/internal/tui/components/footer"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"
)

// DetermineMode determines the appropriate footer mode based on UI state
func DetermineMode(s *ViewState) footer.Mode {
	if s.UI.InputActive {
		switch s.InputMode {
		case state.InputSearch:
			return footer.ModeSearching
		case state.InputRename:
			return footer.ModeRenaming
		case state.InputGoto:
			return footer.ModeGoto
		case state.InputAuth:
			return footer.ModeAuth
		}
	}
	if s.UI.HostConfirm {
		return footer.ModeHostConfirm
	}
	if s.UI.Confirming {
		return footer.ModeConfirming
	}
	if s.Progress.Visible {
		return footer.ModeProgress
	}
	if s.Msg != "" {
		return footer.ModeMessage
	}
	if s.UI.SettingsOpen {
		return footer.ModeSettings
	}
	return footer.ModeNormal
}

func memoizePrompts(s *ViewState, styles theme.Stylesheet) {
	if s.UI.Confirming {
		clipboardCount := len(s.Clipboard.Paths)
		key := fmt.Sprintf("confirm-%s-%d-%s", s.ActionType, clipboardCount, s.ConflictDst)
		if _, ok := s.UI.PromptCache[key]; !ok {
			// Clear old confirm prompts from cache
			for k := range s.UI.PromptCache {
				if strings.HasPrefix(k, "confirm-") {
					delete(s.UI.PromptCache, k)
				}
			}
			prompt := footer.BuildConfirmationPrompt(footer.Props{
				ActionType:     s.ActionType,
				ClipboardCount: clipboardCount,
				ConflictDst:    s.ConflictDst,
			})
			s.UI.PromptCache[key] = footer.ColorizeKeys(footer.Props{Styles: styles}, prompt)
		}
	}

	if s.UI.HostConfirm {
		hostname := ""
		if s.HostConfirmReq != nil {
			hostname = s.HostConfirmReq.Hostname
		}
		key := "hostconfirm-" + hostname
		if _, ok := s.UI.PromptCache[key]; !ok {
			// Clear old hostconfirm prompts
			for k := range s.UI.PromptCache {
				if strings.HasPrefix(k, "hostconfirm-") {
					delete(s.UI.PromptCache, k)
				}
			}
			prompt := fmt.Sprintf("Add host '%s' to known_hosts? (y/n)", hostname)
			s.UI.PromptCache[key] = footer.ColorizeKeys(footer.Props{Styles: styles}, prompt)
		}
	}
}
