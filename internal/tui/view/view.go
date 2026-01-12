package view

import (
	"fmt"
	"strings"

	"fm/internal/config"
	"fm/internal/constants"
	"fm/internal/files"
	"fm/internal/files/sorting"
	"fm/internal/sshutil"
	"fm/internal/tui/components/footer"
	"fm/internal/tui/components/header"
	"fm/internal/tui/state"
	"fm/internal/tui/theme"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ViewState contains all data needed to render the UI.
type ViewState struct {
	Width           int
	Height          int
	Path            string
	Items           []files.Item
	FilteredItems   []files.Item
	Cursor          int
	Offset          int
	SortMode        sorting.SortMode
	GitBranch       string
	GitRoot         string
	ReadOnly        bool
	UI              *state.UIState
	Progress        *state.ProgressState
	Clipboard       *state.ClipboardState
	ProcessingItems map[string]bool
	SelectedPaths   map[string]bool
	SettingsCursor  int
	SettingsOffset  int
	Config          *config.Config
	ActionType      constants.ActionType
	ConflictSrc     string
	ConflictDst     string
	PendingItems    []string
	ActiveInput     textinput.Model
	InputMode       state.InputMode
	HostConfirmReq  *sshutil.HostConfirmRequest
	Msg             string
	Tabs            []state.Tab
	ActiveTab       int
	Separator       string
	LoadingSpinner  spinner.Model
	ViewportHeight  int
	SelectedCount   int
}

// GetViewState constructs the current view state from the model.
func GetViewState(m *state.Model) ViewState {
	return ViewState{
		Width:           m.Display.Width,
		Height:          m.Display.Height,
		ViewportHeight:  m.Display.ViewportHeight,
		Path:            m.Navigation.Path,
		Items:           m.Navigation.Items,
		FilteredItems:   m.Navigation.FilteredItems,
		Cursor:          m.Navigation.Cursor,
		Offset:          m.Navigation.Offset,
		SortMode:        m.Display.SortMode,
		GitBranch:       m.Git.Branch,
		GitRoot:         m.Git.Root,
		ReadOnly:        m.Display.ReadOnly,
		UI:              &m.UI,
		Progress:        &m.Operations.Progress,
		Clipboard:       &m.Operations.Clipboard,
		ProcessingItems: m.Operations.ProcessingItems,
		SelectedPaths:   m.Operations.SelectedPaths,
		SettingsCursor:  m.Settings.Cursor,
		SettingsOffset:  m.Settings.Offset,
		Config:          &m.Config,
		ActionType:      m.Operations.ActionType,
		ConflictSrc:     m.Operations.Conflict.Source,
		ConflictDst:     m.Operations.Conflict.Destination,
		PendingItems:    m.Operations.Conflict.PendingItems,
		ActiveInput:     m.Inputs.ActiveInput,
		InputMode:       m.Inputs.Mode,
		HostConfirmReq:  m.Remote.HostConfirmReq,
		Msg:             m.Message.Text,
		Tabs:            m.Tabs,
		ActiveTab:       m.ActiveTab,
		Separator:       m.FS.Separator(),
		LoadingSpinner:  m.Display.LoadingSpinner,
		SelectedCount:   m.Navigation.SelectedCount,
	}
}

// DetermineMode determines the appropriate footer mode based on UI state
func DetermineMode(s *ViewState) footer.Mode {
	if s.Progress.Visible {
		return footer.ModeProgress
	}
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
	if s.Msg != "" {
		return footer.ModeMessage
	}
	if s.UI.SettingsOpen {
		return footer.ModeSettings
	}
	return footer.ModeNormal
}

// Render assembles the full UI from components.
func Render(s *ViewState, styles theme.Stylesheet) string {
	if s.Width == 0 {
		return "Loading..."
	}

	// Memoize prompts if needed
	if s.UI.Confirming || s.UI.HostConfirm {
		memoizePrompts(s, styles)
	}

	headerStr := renderHeader(s, styles)
	footerStr := renderFooter(s, styles)
	bodyStr := renderBody(s, headerStr, footerStr, styles)

	return lipgloss.JoinVertical(lipgloss.Left, headerStr, bodyStr, footerStr)
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

// renderHeader builds props and renders header
func renderHeader(s *ViewState, styles theme.Stylesheet) string {
	return header.Render(header.Props{
		Width:        s.Width,
		Path:         s.Path,
		Separator:    s.Separator,
		GitBranch:    s.GitBranch,
		ReadOnly:     s.ReadOnly,
		TabCount:     len(s.Tabs),
		ActiveTab:    s.ActiveTab,
		SettingsOpen: s.UI.SettingsOpen,
		Styles:       styles,
	})
}

// renderFooter builds props and renders footer
func renderFooter(s *ViewState, styles theme.Stylesheet) string {
	totalItems := len(s.FilteredItems)
	cursor := s.Cursor
	if totalItems > 0 && len(s.FilteredItems) > 0 && s.FilteredItems[0].IsUp {
		totalItems--
		cursor--
	}

	return footer.Render(footer.Props{
		Mode:            DetermineMode(s),
		Width:           s.Width,
		ProgressLabel:   s.Progress.Label,
		ProgressPercent: s.Progress.Percent,
		ActiveInput:     s.ActiveInput,
		Message:         s.Msg,
		SortMode:        s.SortMode,
		Cursor:          cursor,
		TotalItems:      totalItems,
		SelectedCount:   s.SelectedCount,
		Items:           s.Items,
		FilteredItems:   s.FilteredItems,
		SettingsCursor:  s.SettingsCursor,
		ActionType:      s.ActionType,
		ClipboardCount:  len(s.Clipboard.Paths),
		ConflictDst:     s.ConflictDst,
		HostConfirmReq:  s.HostConfirmReq,
		Styles:          styles,
		PromptCache:     s.UI.PromptCache,
	})
}

// renderBody renders the main content area (List or Settings)
func renderBody(s *ViewState, headerStr, footerStr string, styles theme.Stylesheet) string {
	if s.UI.SettingsOpen {
		return RenderSettingsList(s, headerStr, footerStr, styles)
	}
	return RenderList(s, headerStr, footerStr, styles)
}

// GetViewportHeight calculates the available viewport height efficiently
func GetViewportHeight(s *ViewState) int {
	if s.ViewportHeight > 0 {
		return s.ViewportHeight
	}
	return CalculateViewportHeightFromState(s)
}

// CalculateViewportHeight calculates viewport height from Model
func CalculateViewportHeight(m *state.Model) int {
	// App Header: 1 line
	// App Footer: 1 line
	h := m.Display.Height - 2

	// List Header: 3 lines (separator, text, separator)
	if m.Config.ShowHeader && !m.UI.SettingsOpen {
		h -= 3
	}

	if h < 1 {
		return 1
	}
	return h
}

// CalculateViewportHeightFromState calculates viewport height from ViewState
func CalculateViewportHeightFromState(s *ViewState) int {
	h := s.Height - 2
	if s.Config.ShowHeader && !s.UI.SettingsOpen {
		h -= 3
	}
	if h < 1 {
		return 1
	}
	return h
}
