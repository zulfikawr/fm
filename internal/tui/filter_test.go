package tui

import (
	"testing"

	"filemanager/internal/files"
)

func TestFiltering(t *testing.T) {
	m := NewModel(&files.LocalFS{}, "/")
	m.items = []files.Item{
		{Name: "↑ ..", IsDir: true, IsUp: true},
		{Name: "dir1", IsDir: true, Path: "dir1"},
		{Name: "file1", IsDir: false, Path: "file1"},
	}

	// Mock search input
	m.searching = true
	m.searchInput.SetValue("file")
	m.applyFilter()

	// Items: ↑ .., dir1, file1. Filter "file" should show ↑ .. and file1
	if len(m.filteredItems) != 2 {
		t.Errorf("Expected 2 filtered items, got %d", len(m.filteredItems))
	}
	if m.filteredItems[1].Name != "file1" {
		t.Errorf("Expected filtered item to be file1, got %s", m.filteredItems[1].Name)
	}
}
