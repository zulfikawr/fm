package view

import (
	"fm/internal/config"
	"fm/internal/constants"
	"fm/internal/files/core"
	"fm/internal/files/sorting"
	"fm/internal/sshutil"
	"fm/internal/tui/state"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

// ViewState contains all data needed to render the UI.
type ViewState struct {
	Width           int
	Height          int
	Path            string
	Items           []core.Item
	FilteredItems   []core.Item
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
	RemoteConnected bool
	RemoteUser      string
	RemoteHost      string
	AltMode         bool
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
		SelectedPaths:   m.Navigation.SelectedPaths,
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
		RemoteConnected: !m.FS.IsLocal(),
		RemoteUser:      m.Remote.User,
		RemoteHost:      m.Remote.Host,
		AltMode:         m.Inputs.AltMode,
	}
}
