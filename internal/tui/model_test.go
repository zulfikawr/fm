package tui

import (
	"fm/internal/files"
	"testing"
)

func TestNewModel(t *testing.T) {
	m := NewModel(&files.LocalFS{}, "/")
	if m.path != "/" {
		t.Errorf("Expected path /, got %s", m.path)
	}
	if m.searchInput.Placeholder != "type to search" {
		t.Errorf("Expected placeholder 'type to search', got %s", m.searchInput.Placeholder)
	}
}
