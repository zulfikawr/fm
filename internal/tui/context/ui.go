package context

import (
	"github.com/zulfikawr/fm/internal/tui/components/ui"
)

// --- UI State ---

// UIState holds UI mode flags
type UIState struct {
	Confirming      bool
	SettingsOpen    bool
	HelpOpen        bool
	LogOpen         bool
	ClipboardOpen   bool
	Loading         bool
	SelectMode      bool
	InputActive     bool              // Consolidated flag for any text input (search, rename, etc)
	RemoteAuth      bool              // Specific flag for remote auth (uses input)
	HostConfirm     bool              // Waiting for known_hosts confirmation (uses y/n keys)
	TestingIcons    bool              // Icon support test flow
	UpdateAvailable bool              // New version is available
	LatestVersion   string            // Latest version string
	PromptCache     map[string]string // Pre-calculated styled prompts
}

// Reset resets all UI flags to false
func (ui *UIState) Reset() {
	ui.Confirming = false
	ui.SettingsOpen = false
	ui.HelpOpen = false
	ui.LogOpen = false
	ui.ClipboardOpen = false
	ui.Loading = false
	ui.SelectMode = false
	ui.InputActive = false
	ui.RemoteAuth = false
	ui.HostConfirm = false
	ui.TestingIcons = false
	ui.PromptCache = make(map[string]string)
}

// StartInput enters an input mode
func (ui *UIState) StartInput() {
	ui.InputActive = true
	ui.LogOpen = false
	ui.HelpOpen = false
	ui.ClipboardOpen = false
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
	ui.HelpOpen = false
	ui.LogOpen = false
	ui.ClipboardOpen = false
}

// StopConfirming exits confirmation mode
func (ui *UIState) StopConfirming() {
	ui.Confirming = false
}

// ToggleSettings toggles the settings view
func (ui *UIState) ToggleSettings() {
	ui.SettingsOpen = !ui.SettingsOpen
	if ui.SettingsOpen {
		ui.InputActive = false
		ui.Confirming = false
		ui.HelpOpen = false
		ui.LogOpen = false
		ui.ClipboardOpen = false
	}
}

// ToggleHelp toggles the help view
func (ui *UIState) ToggleHelp() {
	ui.HelpOpen = !ui.HelpOpen
	if ui.HelpOpen {
		ui.InputActive = false
		ui.Confirming = false
		ui.SettingsOpen = false
		ui.LogOpen = false
		ui.ClipboardOpen = false
	}
}

// ToggleLogs toggles the log view
func (ui *UIState) ToggleLogs() {
	ui.LogOpen = !ui.LogOpen
	if ui.LogOpen {
		ui.InputActive = false
		ui.Confirming = false
		ui.SettingsOpen = false
		ui.HelpOpen = false
		ui.ClipboardOpen = false
	}
}

// ToggleClipboard toggles the clipboard view
func (ui *UIState) ToggleClipboard() {
	ui.ClipboardOpen = !ui.ClipboardOpen
	if ui.ClipboardOpen {
		ui.InputActive = false
		ui.Confirming = false
		ui.SettingsOpen = false
		ui.HelpOpen = false
		ui.LogOpen = false
	}
}

// --- Input State ---

// InputState holds the unified text input model.
type InputState struct {
	ActiveInput ui.Input  // The single shared text input
	Mode        InputMode // What we are currently inputting
	AltMode     bool      // Toggled alternative mode (e.g., Remote for Goto, KeyPath for Auth)
}
