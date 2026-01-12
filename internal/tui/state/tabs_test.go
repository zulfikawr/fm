package state

import (
	"fm/internal/files/sorting"
	"testing"
)

func TestTab(t *testing.T) {
	t.Run("NewTab", func(t *testing.T) {
		tab := NewTab("/test", sorting.SortName)
		if tab.Path != "/test" || tab.SortMode != sorting.SortName || tab.SelectedPaths == nil {
			t.Errorf("NewTab failed: %+v", tab)
		}
	})

	t.Run("Selection", func(t *testing.T) {
		tab := NewTab("/", sorting.SortDefault)
		tab.SelectedPaths["/p1"] = true

		if tab.SelectedCount() != 1 {
			t.Errorf("SelectedCount failed: %d", tab.SelectedCount())
		}
		if !tab.IsSelected("/p1") {
			t.Error("IsSelected failed for true")
		}
		if tab.IsSelected("/p2") {
			t.Error("IsSelected failed for false")
		}
	})
}
