package context

import (
	"github.com/zulfikawr/fm/internal/tui/components/ui"
)

// ViewMode represents the currently active main view
type ViewMode int

const (
	ViewMain ViewMode = iota
	ViewSettings
	ViewHelp
	ViewLogs
	ViewClipboard
	ViewTrash
	ViewAnalyze
)

// --- UI State ---

// UIState holds UI mode flags and active view state
type UIState struct {
	ActiveView      ViewMode
	Confirming      bool
	Loading         bool
	InputActive     bool              // Consolidated flag for any text input (search, rename, etc)
	RemoteAuth      bool              // Specific flag for remote auth (uses input)
	HostConfirm     bool              // Waiting for known_hosts confirmation (uses y/n keys)
	TestingIcons    bool              // Icon support test flow
	UpdateAvailable bool              // New version is available
	LatestVersion   string            // Latest version string
	PromptCache     map[string]string // Pre-calculated styled prompts
}

// Reset resets all UI flags and returns to main view
func (ui *UIState) Reset() {
	ui.ActiveView = ViewMain
	ui.Confirming = false
	ui.Loading = false
	ui.InputActive = false
	ui.RemoteAuth = false
	ui.HostConfirm = false
	ui.TestingIcons = false
	ui.PromptCache = make(map[string]string)
}

// StartInput enters an input mode
func (ui *UIState) StartInput() {
	ui.InputActive = true
	ui.Confirming = false
}

// StopInput exits input mode
func (ui *UIState) StopInput() {
	ui.InputActive = false
}

// StartConfirming enters confirmation mode
func (ui *UIState) StartConfirming() {
	ui.Confirming = true
	ui.InputActive = false
}

// StopConfirming exits confirmation mode
func (ui *UIState) StopConfirming() {
	ui.Confirming = false
}

// SetView sets the active view, ensuring inputs/confirmations are cleared if switching away from main
func (ui *UIState) SetView(view ViewMode) {
	if view != ViewMain {
		ui.InputActive = false
		ui.Confirming = false
	}
	ui.ActiveView = view
}

// ToggleSettings toggles the settings view
func (ui *UIState) ToggleSettings() {
	if ui.ActiveView == ViewSettings {
		ui.ActiveView = ViewMain
	} else {
		ui.SetView(ViewSettings)
	}
}

// ToggleHelp toggles the help view
func (ui *UIState) ToggleHelp() {
	if ui.ActiveView == ViewHelp {
		ui.ActiveView = ViewMain
	} else {
		ui.SetView(ViewHelp)
	}
}

// ToggleLogs toggles the log view
func (ui *UIState) ToggleLogs() {
	if ui.ActiveView == ViewLogs {
		ui.ActiveView = ViewMain
	} else {
		ui.SetView(ViewLogs)
	}
}

// ToggleClipboard toggles the clipboard view
func (ui *UIState) ToggleClipboard() {
	if ui.ActiveView == ViewClipboard {
		ui.ActiveView = ViewMain
	} else {
		ui.SetView(ViewClipboard)
	}
}

// ToggleTrash toggles the trash view
func (ui *UIState) ToggleTrash() {
	if ui.ActiveView == ViewTrash {
		ui.ActiveView = ViewMain
	} else {
		ui.SetView(ViewTrash)
	}
}

// --- Input State ---

// InputState holds the unified text input model.
type InputState struct {
	ActiveInput ui.Input  // The single shared text input
	Mode        InputMode // What we are currently inputting
	AltMode     bool      // Toggled alternative mode (e.g., Remote for Goto, KeyPath for Auth)
}
