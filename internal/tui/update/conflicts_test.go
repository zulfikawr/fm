package update

import (
	tuitestutil "fm/internal/tui/testutil"
	"testing"

	"fm/internal/tui/commands"
)

func TestHandleConflict(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	msg := commands.ConflictMsg{
		Src:          "src",
		Dst:          "dst",
		PendingItems: []string{"p1"},
		IsMove:       true,
	}

	HandleConflict(m, msg)
	if !m.UI.Confirming {
		t.Error("Expected confirming to be true")
	}
	if m.Operations.Conflict.Source != "src" {
		t.Errorf("Expected src, got %s", m.Operations.Conflict.Source)
	}
}
