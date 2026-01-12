package actions

import (
	"fm/internal/logger"
	"fm/internal/tui/commands"
	tuierrors "fm/internal/tui/errors"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrorHandler provides centralized error handling for the TUI
var ErrorHandler = &tuierrors.Handler{
	OnUser: func(err *tuierrors.Error) error {
		// User errors are displayed in the UI
		logger.Info(err.LogMessage())
		return err
	},
	OnSystem: func(err *tuierrors.Error) error {
		// System errors are logged and shown as generic message to user
		logger.Error(err.LogMessage())
		return err
	},
	OnFatal: func(err *tuierrors.Error) error {
		// Fatal errors are logged with stack trace
		logger.Error(err.LogMessage())
		logger.Error(err.StackTrace())
		return err
	},
	OnTransient: func(err *tuierrors.Error) error {
		// Transient errors can be retried
		if err.ShouldRetry() {
			logger.Info(err.LogMessage())
		} else {
			logger.Error(err.LogMessage())
		}
		return err
	},
	Logger: func(err *tuierrors.Error) {
		// Log all errors for debugging (using Info level since Debug doesn't exist)
		logger.Info(err.LogMessage())
	},
}

// LogError logs a TUI error and sets it as the current error in the model
func LogError(m *state.Model, err error, context string) tea.Cmd {
	if err == nil {
		return nil
	}

	// Convert to TUI error if needed
	var tuiErr *tuierrors.Error
	var ok bool
	if tuiErr, ok = err.(*tuierrors.Error); !ok {
		// Wrap standard errors as system errors
		tuiErr = tuierrors.SystemError(context, err)
	}

	// Handle the error through our handler
	ErrorHandler.Handle(tuiErr)

	// Set in model for display
	m.Message.Error = tuiErr

	// Display appropriate message in UI
	userMsg := tuiErr.UserMessage()
	return commands.SetMsg(m, "Error: "+userMsg)
}

// LogInfo logs an informational message
func LogInfo(m *state.Model, msg string) tea.Cmd {
	logger.Info(msg)
	return commands.SetMsg(m, msg)
}

// Reload triggers an asynchronous reload of the current directory
func Reload(m *state.Model) tea.Cmd {
	m.UI.Loading = true
	return commands.Reload(m.FS, m.GS, m.Navigation.Path, m.Navigation.PathGen, m.Display.SortMode, m.Config.ShowHidden)
}

// SaveTabState saves the current model state to the active tab
func SaveTabState(m *state.Model) {
	if m.ActiveTab >= 0 && m.ActiveTab < len(m.Tabs) {
		m.Tabs[m.ActiveTab].Path = m.Navigation.Path
		m.Tabs[m.ActiveTab].Items = m.Navigation.Items
		m.Tabs[m.ActiveTab].FilteredItems = m.Navigation.FilteredItems
		m.Tabs[m.ActiveTab].Cursor = m.Navigation.Cursor
		m.Tabs[m.ActiveTab].Offset = m.Navigation.Offset
		m.Tabs[m.ActiveTab].SortMode = m.Display.SortMode
		m.Tabs[m.ActiveTab].GitBranch = m.Git.Branch
		m.Tabs[m.ActiveTab].GitRoot = m.Git.Root
		m.Tabs[m.ActiveTab].Searching = m.UI.InputActive && m.Inputs.Mode == state.InputSearch
		m.Tabs[m.ActiveTab].SearchQuery = m.Inputs.ActiveInput.Value()
		m.Tabs[m.ActiveTab].SelectMode = m.UI.SelectMode
		m.Tabs[m.ActiveTab].SelectedPaths = make(map[string]bool)
		for k, v := range m.Operations.SelectedPaths {
			m.Tabs[m.ActiveTab].SelectedPaths[k] = v
		}
	}
}

// SyncTabToModel loads the active tab's state into the model
func SyncTabToModel(m *state.Model) {
	if m.ActiveTab >= 0 && m.ActiveTab < len(m.Tabs) {
		tab := m.Tabs[m.ActiveTab]
		m.Navigation.Path = tab.Path
		m.Navigation.Items = tab.Items
		m.Navigation.FilteredItems = tab.FilteredItems
		m.Navigation.Cursor = tab.Cursor
		m.Navigation.Offset = tab.Offset
		m.Display.SortMode = tab.SortMode
		m.Git.Branch = tab.GitBranch
		m.Git.Root = tab.GitRoot

		if tab.Searching {
			m.UI.InputActive = true
			m.Inputs.Mode = state.InputSearch
			m.Inputs.ActiveInput.Focus()
			m.Inputs.ActiveInput.Prompt = "/ "
		} else {
			m.UI.InputActive = false
			m.Inputs.Mode = state.InputNone
			m.Inputs.ActiveInput.Blur()
		}

		m.Inputs.ActiveInput.SetValue(tab.SearchQuery)
		m.UI.SelectMode = tab.SelectMode
		m.Operations.SelectedPaths = make(map[string]bool)
		for k, v := range tab.SelectedPaths {
			m.Operations.SelectedPaths[k] = v
		}
		m.Navigation.PathGen++
	}
}
