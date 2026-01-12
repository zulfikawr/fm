package commands

import (
	"errors"
	"testing"

	"fm/internal/files"
)

func TestMessages(t *testing.T) {
	t.Run("LoadedItemsMsg", func(t *testing.T) {
		msg := LoadedItemsMsg{
			Generation: 1,
			Path:       "/test",
			Items:      []files.Item{{Name: "test"}},
			Err:        errors.New("test error"),
		}
		if msg.Path != "/test" || len(msg.Items) != 1 || msg.Err == nil {
			t.Errorf("LoadedItemsMsg not initialized correctly: %+v", msg)
		}
	})

	t.Run("GitStatusMsg", func(t *testing.T) {
		msg := GitStatusMsg{
			Path:     "/test",
			Statuses: map[string]string{"file": "M"},
			Branch:   "main",
		}
		if msg.Branch != "main" || msg.Statuses["file"] != "M" {
			t.Errorf("GitStatusMsg not initialized correctly: %+v", msg)
		}
	})

	t.Run("ErrorMsg", func(t *testing.T) {
		err := errors.New("fail")
		msg := ErrorMsg{Err: err}
		if msg.Err != err {
			t.Errorf("ErrorMsg not initialized correctly: %+v", msg)
		}
	})

	t.Run("ConflictMsg", func(t *testing.T) {
		msg := ConflictMsg{
			Src:    "src",
			Dst:    "dst",
			IsMove: true,
		}
		if msg.Src != "src" || !msg.IsMove {
			t.Errorf("ConflictMsg not initialized correctly: %+v", msg)
		}
	})

	t.Run("OperationFinishedMsg", func(t *testing.T) {
		msg := OperationFinishedMsg{Paths: []string{"p1", "p2"}}
		if len(msg.Paths) != 2 {
			t.Errorf("OperationFinishedMsg not initialized correctly: %+v", msg)
		}
	})
}
