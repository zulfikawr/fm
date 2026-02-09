package view_test

import (
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
	"github.com/zulfikawr/fm/internal/tui/context"
	"github.com/zulfikawr/fm/internal/tui/view"
)

func TestCalculateLayout(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	m := context.NewModel(fs, "/test")
	m.Config.UI.ShowHeader = false

	m.Display.Width = 100
	m.Display.Height = 20
	m.SyncViewportHeight()

	layout := view.CalculateLayout(m)
	testutil.AssertEqual(t, 100, layout.Width, "Width should match")
	testutil.AssertEqual(t, 20, layout.Height, "Height should match")
	testutil.AssertEqual(t, 1, layout.HeaderHeight, "Header height should be 1")
	testutil.AssertEqual(t, 1, layout.FooterHeight, "Footer height should be 1")
	testutil.AssertEqual(t, 18, layout.BodyHeight, "Body height should be 18")

	testutil.AssertEqual(t, 18, view.GetViewportHeight(m), "Viewport height should match body height")

	t.Run("Header enabled", func(t *testing.T) {
		m.Config.UI.ShowHeader = true
		m.SyncViewportHeight()
		testutil.AssertEqual(t, 15, view.GetViewportHeight(m), "Viewport height should be 15 when header is enabled")
	})

	t.Run("Zero height", func(t *testing.T) {
		m.Display.Height = 0
		m.SyncViewportHeight()
		layout := view.CalculateLayout(m)
		testutil.AssertEqual(t, 0, layout.BodyHeight, "Body height should be 0")
	})

	t.Run("Small height", func(t *testing.T) {
		m.Display.Height = 1
		layout := view.CalculateLayout(m)
		testutil.AssertEqual(t, 0, layout.BodyHeight, "Body height should be 0 when height is 1")
	})
}
