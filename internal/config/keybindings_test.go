package config

import (
	"strconv"
	"testing"
)

func TestKeybindings_Validation(t *testing.T) {
	t.Run("Default keybindings", func(t *testing.T) {
		if err := ValidateKeybindings(DefaultKeybindings()); err != nil {
			t.Fatalf("default keybindings should be valid: %v", err)
		}
	})

	t.Run("Conflict detection", func(t *testing.T) {
		keybinds := []Keybinding{
			{Action: "quit", Keys: []string{"q"}},
			{Action: "open", Keys: []string{"q"}},
		}
		err := ValidateKeybindings(keybinds)
		if err == nil {
			t.Error("Expected error for conflicting keybindings, got nil")
		}
	})

	t.Run("Browser hijack detection (warning only)", func(t *testing.T) {
		keybinds := []Keybinding{
			{Action: "quit", Keys: []string{"ctrl+t"}},
		}
		err := ValidateKeybindings(keybinds)
		if err != nil {
			t.Errorf("Expected no error (warning only) for browser hijacked key, got %v", err)
		}
	})

	t.Run("Valid combination", func(t *testing.T) {
		keybinds := []Keybinding{
			{Action: "quit", Keys: []string{"ctrl+c", "q"}},
			{Action: "open", Keys: []string{"enter", "l"}},
		}
		err := ValidateKeybindings(keybinds)
		if err != nil {
			t.Errorf("Expected no error for valid bindings, got %v", err)
		}
	})

	t.Run("Same key for same action is allowed", func(t *testing.T) {
		keybinds := []Keybinding{
			{Action: "quit", Keys: []string{"q", "q"}},
		}
		err := ValidateKeybindings(keybinds)
		if err != nil {
			t.Errorf("Expected no error for duplicate keys in same action, got %v", err)
		}
	})
}

func TestDefaultKeybindings(t *testing.T) {
	want := map[string]string{
		"settings":            ",",
		"select_all":          "ctrl+a",
		"new_tab":             "ctrl+t",
		"close_tab":           "ctrl+w",
		"create":              "ctrl+n",
		"fuzzy_search":        "ctrl+f",
		"toggle_regex_search": "ctrl+r",
		"analyze":             "ctrl+u",
		"clipboard_view":      "ctrl+b",
		"logs_view":           "ctrl+l",
	}

	keybindings := DefaultKeybindings()
	for action, key := range want {
		keys := GetKeybindingForAction(action, keybindings)
		if len(keys) != 1 || keys[0] != key {
			t.Errorf("%s: got %v, want [%s]", action, keys, key)
		}
	}

	for tab := 1; tab <= 9; tab++ {
		key := strconv.Itoa(tab)
		action := "switch_tab_" + key
		keys := GetKeybindingForAction(action, keybindings)
		if len(keys) != 1 || keys[0] != key {
			t.Errorf("%s: got %v, want [%s]", action, keys, key)
		}
	}
}

func TestKeybindings_Helpers(t *testing.T) {
	keybinds := []Keybinding{
		{Action: "quit", Keys: []string{"ctrl+c", "q"}},
		{Action: "open", Keys: []string{"enter"}},
	}

	t.Run("GetActionForKey", func(t *testing.T) {
		if GetActionForKey("q", keybinds) != "quit" {
			t.Error("Failed to find action for 'q'")
		}
		if GetActionForKey("ENTER", keybinds) != "open" {
			t.Error("Failed to find action for 'ENTER' (case-insensitive check)")
		}
		if GetActionForKey("unknown", keybinds) != "" {
			t.Error("Found action for unknown key")
		}
	})

	t.Run("GetKeybindingForAction", func(t *testing.T) {
		keys := GetKeybindingForAction("quit", keybinds)
		if len(keys) != 2 || keys[1] != "q" {
			t.Errorf("Expected [ctrl+c, q], got %v", keys)
		}
	})
}

func TestKeybindings_HumanLabel(t *testing.T) {
	t.Run("Special labels", func(t *testing.T) {
		kb := Keybinding{Action: "go_parent"}
		if kb.HumanLabel() != "Go to Parent" {
			t.Errorf("Expected 'Go to Parent', got %q", kb.HumanLabel())
		}
	})

	t.Run("Fallback title case", func(t *testing.T) {
		kb := Keybinding{Action: "custom_action_name"}
		if kb.HumanLabel() != "Custom Action Name" {
			t.Errorf("Expected 'Custom Action Name', got %q", kb.HumanLabel())
		}
	})
}
