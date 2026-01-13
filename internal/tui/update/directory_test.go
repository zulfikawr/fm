package update

import (
	"testing"

	"fm/internal/files/core"
	"fm/internal/tui/commands"
	tuitestutil "fm/internal/tui/testutil"
)

func TestHandleLoadedItems(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	msg := commands.LoadedItemsMsg{
		Generation: 0,
		Path:       "/test",
		Items:      []core.Item{{Name: "f1"}},
	}

	_, handled := HandleLoadedItems(m, msg)
	if handled {
		t.Error("Expected handled false for successful load routing")
	}
	if len(m.Navigation.Items) != 1 {
		t.Error("Expected 1 item in model")
	}
}
