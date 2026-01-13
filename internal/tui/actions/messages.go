package actions

import (
	"fm/internal/logger"
	"fm/internal/tui/commands"
	tuierrors "fm/internal/tui/errors"
	"fm/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

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
