package config

import (
	"fmt"
	"strings"

	"github.com/zulfikawr/fm/internal/logger"
)

// Keybinding holds a single keybinding configuration
type Keybinding struct {
	Action   string   `json:"action"`
	Keys     []string `json:"keys"`
	Category string   `json:"category"` // navigation, file_ops, tabs, etc.
}

// KeybindingConfig holds all keybinding configurations
type KeybindingConfig struct {
	Version   int                 `json:"version"`
	Keybinds  []Keybinding        `json:"keybindings"`
	Overrides map[string][]string `json:"overrides,omitempty"` // action -> [keys]
}

const CurrentKeybindingVersion = 1

// DefaultKeybindings returns the default keybinding configuration
func DefaultKeybindings() []Keybinding {
	return []Keybinding{
		// Navigation
		{Action: "open", Keys: []string{"enter", "l", "right"}, Category: "navigation"},
		{Action: "go_parent", Keys: []string{"backspace", "h", "left"}, Category: "navigation"},
		{Action: "move_down", Keys: []string{"j", "down"}, Category: "navigation"},
		{Action: "move_up", Keys: []string{"k", "up"}, Category: "navigation"},
		{Action: "page_down", Keys: []string{"pgdown"}, Category: "navigation"},
		{Action: "page_up", Keys: []string{"pgup"}, Category: "navigation"},

		// Selection
		{Action: "toggle_selection", Keys: []string{" "}, Category: "selection"},
		{Action: "select_all", Keys: []string{"alt+a"}, Category: "selection"},
		{Action: "clear_selection", Keys: []string{"esc"}, Category: "selection"},

		// Tabs
		{Action: "new_tab", Keys: []string{"alt+t"}, Category: "tabs"},
		{Action: "close_tab", Keys: []string{"alt+w"}, Category: "tabs"},
		{Action: "switch_tab_1", Keys: []string{"alt+1"}, Category: "tabs"},
		{Action: "switch_tab_2", Keys: []string{"alt+2"}, Category: "tabs"},
		{Action: "switch_tab_3", Keys: []string{"alt+3"}, Category: "tabs"},
		{Action: "switch_tab_4", Keys: []string{"alt+4"}, Category: "tabs"},
		{Action: "switch_tab_5", Keys: []string{"alt+5"}, Category: "tabs"},
		{Action: "switch_tab_6", Keys: []string{"alt+6"}, Category: "tabs"},
		{Action: "switch_tab_7", Keys: []string{"alt+7"}, Category: "tabs"},
		{Action: "switch_tab_8", Keys: []string{"alt+8"}, Category: "tabs"},
		{Action: "switch_tab_9", Keys: []string{"alt+9"}, Category: "tabs"},

		// File Operations
		{Action: "copy", Keys: []string{"c"}, Category: "file_ops"},
		{Action: "cut", Keys: []string{"x"}, Category: "file_ops"},
		{Action: "paste", Keys: []string{"v"}, Category: "file_ops"},
		{Action: "rename", Keys: []string{"r"}, Category: "file_ops"},
		{Action: "delete", Keys: []string{"d"}, Category: "file_ops"},
		{Action: "create", Keys: []string{"alt+n"}, Category: "file_ops"},
		{Action: "zip", Keys: []string{"z"}, Category: "file_ops"},
		{Action: "unzip", Keys: []string{"u"}, Category: "file_ops"},

		// Search & Filter
		{Action: "filter", Keys: []string{"/"}, Category: "search"},
		{Action: "fuzzy_search", Keys: []string{"alt+/"}, Category: "search"},
		{Action: "toggle_regex_search", Keys: []string{"alt+r"}, Category: "search"},

		{Action: "quit", Keys: []string{"ctrl+c"}, Category: "general"},
		{Action: "settings", Keys: []string{"."}, Category: "general"},
		{Action: "help", Keys: []string{"?"}, Category: "general"},
		{Action: "analyze", Keys: []string{"alt+u"}, Category: "general"},
		{Action: "trash_view", Keys: []string{"t"}, Category: "general"},
		{Action: "clipboard_view", Keys: []string{"alt+c"}, Category: "general"},
		{Action: "logs_view", Keys: []string{"alt+l"}, Category: "general"},
		{Action: "go_to_path", Keys: []string{"g"}, Category: "navigation"},
		{Action: "history_back", Keys: []string{"["}, Category: "navigation"},
		{Action: "history_forward", Keys: []string{"]"}, Category: "navigation"},
		{Action: "cycle_sort", Keys: []string{"s"}, Category: "navigation"},
	}
}

// ValidateKeybindings checks for conflicts and invalid combinations
func ValidateKeybindings(keybinds []Keybinding) error {
	keyToActions := make(map[string][]string)

	// Collect all key-action mappings
	for i := range keybinds {
		kb := keybinds[i]
		for j := range kb.Keys {
			key := strings.ToLower(strings.TrimSpace(kb.Keys[j]))
			if existingActions, exists := keyToActions[key]; exists {
				// Check if this is the same action (allowed for multiple keys per action)
				isSameAction := false
				for k := range existingActions {
					if existingActions[k] == kb.Action {
						isSameAction = true
						break
					}
				}
				if !isSameAction {
					return fmt.Errorf("key '%s' is bound to multiple actions: [%s, %s]",
						key, strings.Join(existingActions, ", "), kb.Action)
				}
			}
			keyToActions[key] = append(keyToActions[key], kb.Action)
		}
	}

	// Check for potentially problematic browser hijacked keys
	browserHijacked := map[string]string{
		"ctrl+t":       "opens new browser tab",
		"ctrl+w":       "closes browser tab",
		"ctrl+n":       "opens new browser window",
		"ctrl+shift+t": "reopens closed browser tab",
		"ctrl+shift+n": "opens new incognito window",
		"ctrl+f4":      "closes tab in some browsers",
	}

	for key := range keyToActions {
		if reason, hijacked := browserHijacked[key]; hijacked {
			logger.Warnf("WARNING: Key '%s' is commonly hijacked by browsers (%s)", key, reason)
		}
	}

	return nil
}

// HumanLabel returns a human-readable label for the action
func (kb Keybinding) HumanLabel() string {
	special := map[string]string{
		"go_parent":           "Go to Parent",
		"go_to_path":          "Go to Path",
		"new_tab":             "New Tab",
		"close_tab":           "Close Tab",
		"fuzzy_search":        "Fuzzy Search",
		"toggle_regex_search": "Toggle Regex Search",
		"cycle_sort":          "Cycle Sort",
		"clipboard_view":      "Toggle Clipboard",
		"logs_view":           "Toggle Logs",
		"open":                "Open / Enter",
		"history_back":        "History Back",
		"history_forward":     "History Forward",
		"toggle_selection":    "Toggle Selection",
		"clear_selection":     "Clear Selection",
		"move_up":             "Move Up",
		"move_down":           "Move Down",
		"page_up":             "Page Up",
		"page_down":           "Page Down",
		"file_ops":            "File Operations",
	}

	if human, ok := special[kb.Action]; ok {
		return human
	}

	// Default: replace underscores and title case
	parts := strings.Split(kb.Action, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// GetKeybindingForAction returns the keys associated with a specific action
func GetKeybindingForAction(action string, keybinds []Keybinding) []string {
	for i := range keybinds {
		kb := keybinds[i]
		if kb.Action == action {
			return kb.Keys
		}
	}
	return []string{} // Return empty slice if not found
}

// GetActionForKey returns the action associated with a specific key
func GetActionForKey(key string, keybinds []Keybinding) string {
	key = strings.ToLower(strings.TrimSpace(key))
	for i := range keybinds {
		kb := keybinds[i]
		for j := range kb.Keys {
			if strings.ToLower(kb.Keys[j]) == key {
				return kb.Action
			}
		}
	}
	return "" // Return empty if not found
}
