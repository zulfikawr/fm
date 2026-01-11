package tui

import (
	"strings"
	"testing"

	"fm/internal/files"
)

func TestView(t *testing.T) {
	m := NewModel(&files.LocalFS{}, "/")
	m.width = 80
	m.height = 24

	t.Run("Basic Rendering", func(t *testing.T) {
		view := m.View()
		if view == "" {
			t.Error("View should not be empty")
		}
	})

	t.Run("Loading View", func(t *testing.T) {
		m.loading = true
		m.filteredItems = []files.Item{}
		view := m.View()
		if !strings.Contains(view, "Loading...") {
			t.Error("Expected Loading... in view")
		}
		m.loading = false
	})
}
