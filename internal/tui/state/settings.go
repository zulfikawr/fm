package state

// SettingsState holds settings view state
type SettingsState struct {
	Cursor int // Index in the settings list
	Offset int // Scroll offset for settings
}
