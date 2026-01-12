package tui

import (
	"fm/internal/files/local"
	"fm/internal/tui/theme"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel(&local.LocalFS{}, "/")
	if m.Navigation.Path != "/" {
		t.Errorf("Expected path /, got %s", m.Navigation.Path)
	}
	if m.Inputs.ActiveInput.Placeholder != "" { // it is empty by default now, set in OpenPrompt
		t.Errorf("Expected empty placeholder, got %s", m.Inputs.ActiveInput.Placeholder)
	}
}

func TestThemeApplication(t *testing.T) {
	for i, thm := range theme.Themes {
		t.Run(thm.Name, func(t *testing.T) {
			styles := theme.NewStylesheet(thm)
			// Check some basic style properties
			if styles.Header.GetForeground() != thm.Dir {
				t.Errorf("Expected header foreground %v, got %v", thm.Dir, styles.Header.GetForeground())
			}

			a := NewApp(&local.LocalFS{}, "/")
			m := a.Model
			m.Config.ThemeIndex = i
			m.Display.Width = 80

			s := a.ViewState()
			// Test that view state can be used for rendering
			if s.Width != 80 {
				t.Errorf("Expected width 80, got %d", s.Width)
			}
			if s.Path != "/" {
				t.Errorf("Expected path /, got %s", s.Path)
			}
		})
	}
}

func TestUpdateWindowSize(t *testing.T) {
	a := NewApp(&local.LocalFS{}, "/")
	a.UI.InputActive = false
	newModel, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	a = newModel.(*App)
	if a.Display.Width != 100 || a.Display.Height != 50 {
		t.Errorf("Expected 100x50, got %dx%d", a.Display.Width, a.Display.Height)
	}
}

func TestUpdateQuit(t *testing.T) {
	a := NewApp(&local.LocalFS{}, "/")
	_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("Expected quit command")
	}
}
