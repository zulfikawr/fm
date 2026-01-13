package state

import "github.com/charmbracelet/bubbles/textinput"

// InputMode represents what the current active input is for.
type InputMode int

const (
	InputNone InputMode = iota
	InputSearch
	InputRename
	InputGoto
	InputAuth
)

// InputState holds the unified text input model.
type InputState struct {
	ActiveInput textinput.Model // The single shared text input
	Mode        InputMode       // What we are currently inputting
	AltMode     bool            // Toggled alternative mode (e.g., Remote for Goto, KeyPath for Auth)
}
