package state

import "testing"

func TestProgressState(t *testing.T) {
	p := &ProgressState{}

	t.Run("Show", func(t *testing.T) {
		p.Show("test")
		if !p.Visible || p.Label != "test" || p.Percent != 0 {
			t.Errorf("Show failed: %+v", p)
		}
	})

	t.Run("Update", func(t *testing.T) {
		p.Update(0.5)
		if p.Percent != 0.5 {
			t.Errorf("Update failed: %+v", p)
		}
	})

	t.Run("Hide", func(t *testing.T) {
		p.Hide()
		if p.Visible || p.Label != "" || p.Percent != 0 {
			t.Errorf("Hide failed: %+v", p)
		}
	})
}
