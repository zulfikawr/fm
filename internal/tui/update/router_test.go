package update

import (
	"errors"
	"testing"

	"fm/internal/tui/commands"
	tuitestutil "fm/internal/tui/testutil"
)

func TestHandleGenericMsg(t *testing.T) {
	m := tuitestutil.CreateTestModel()
	m.Message.Text = "hello"

	msg := commands.ClearMsg{}
	_ = HandleGenericMsg(m, msg)
	if m.Message.Text != "" {
		t.Error("Expected message to be cleared")
	}

	// Test ErrorMsg
	err := errors.New("test")
	m.Operations.ProcessingItems = map[string]bool{"f1": true}
	_ = HandleGenericMsg(m, commands.ErrorMsg{Err: err})
	if len(m.Operations.ProcessingItems) != 0 {
		t.Error("Expected processing items to be cleared on error")
	}
}
