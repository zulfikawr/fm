package state

import "testing"

func TestClipboardState(t *testing.T) {
	c := &ClipboardState{}

	t.Run("SetCopy", func(t *testing.T) {
		paths := []string{"p1", "p2"}
		c.SetCopy(nil, paths)
		if len(c.Paths) != 2 || c.IsCut || c.Action != "copy" {
			t.Errorf("SetCopy failed: %+v", c)
		}
	})

	t.Run("SetCut", func(t *testing.T) {
		paths := []string{"p3"}
		c.SetCut(nil, paths)
		if len(c.Paths) != 1 || !c.IsCut || c.Action != "cut" {
			t.Errorf("SetCut failed: %+v", c)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		c.SetCopy(nil, []string{"p"})
		c.Clear()
		if c.Paths != nil || c.IsCut || c.Action != "" {
			t.Errorf("Clear failed: %+v", c)
		}
	})
}
