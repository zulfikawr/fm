package state

// UIState holds UI mode flags
type UIState struct {
	Confirming   bool
	SettingsOpen bool
	Loading      bool
	SelectMode   bool
	InputActive  bool              // Consolidated flag for any text input (search, rename, etc)
	RemoteAuth   bool              // Specific flag for remote auth (uses input)
	HostConfirm  bool              // Waiting for known_hosts confirmation (uses y/n keys)
	PromptCache  map[string]string // Pre-calculated styled prompts
}

// Reset resets all UI flags to false
func (s *UIState) Reset() {
	s.Confirming = false
	s.SettingsOpen = false
	s.Loading = false
	s.SelectMode = false
	s.InputActive = false
	s.RemoteAuth = false
	s.HostConfirm = false
	s.PromptCache = make(map[string]string)
}

// StartInput enters an input mode
func (s *UIState) StartInput() {
	s.InputActive = true
	s.SettingsOpen = false
	s.Confirming = false
}

// StopInput exits input mode
func (s *UIState) StopInput() {
	s.InputActive = false
}

// StartConfirming enters confirmation mode
func (s *UIState) StartConfirming() {
	s.Confirming = true
	s.InputActive = false
	s.SettingsOpen = false
}

// StopConfirming exits confirmation mode
func (s *UIState) StopConfirming() {
	s.Confirming = false
}

// ToggleSettings toggles the settings view
func (s *UIState) ToggleSettings() {
	s.SettingsOpen = !s.SettingsOpen
	if s.SettingsOpen {
		s.InputActive = false
		s.Confirming = false
	}
}
