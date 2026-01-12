package state

import (
	"testing"
)

func TestModel(t *testing.T) {
	m := &Model{
		Navigation: NavigationState{
			Path: "/",
		},
	}
	if m.Navigation.Path != "/" {
		t.Errorf("Expected path /, got %s", m.Navigation.Path)
	}
}
