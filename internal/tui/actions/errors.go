package actions

import (
	"fm/internal/logger"
	tuierrors "fm/internal/tui/errors"
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
