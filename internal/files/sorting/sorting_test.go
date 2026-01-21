package sorting

import (
	"testing"
	"time"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestSortModes(t *testing.T) {
	t.Run("Next cycle", func(t *testing.T) {
		mode := SortDefault
		// Default(0)->Name(1)->NameDesc(2)->Newest(3)->Oldest(4)->SizeDesc(5)->SizeAsc(6)->Default(0)
		mode = mode.Next() // 1
		mode = mode.Next() // 2
		testutil.AssertEqual(t, SortNameDesc, mode, "Should be NameDesc after 2 steps")
	})

	t.Run("String representation", func(t *testing.T) {
		testCases := []struct {
			mode     SortMode
			expected string
		}{
			{SortDefault, "[ ⇅ ] Default"},
			{SortName, "[ A-Z ] Name (Asc)"},
			{SortNameDesc, "[ Z-A ] Name (Desc)"},
			{SortNewest, "[ ↓ ] Newest"},
			{SortOldest, "[ ↑ ] Oldest"},
			{SortSizeDesc, "[ ▼ ] Size (Lrg)"},
			{SortSizeAsc, "[ ▲ ] Size (Sml)"},
			{SortMode(99), "[ ? ] Unknown"},
		}

		for _, tc := range testCases {
			testutil.AssertEqual(t, tc.expected, tc.mode.String(), "String should match")
		}
	})
}

func TestSortItems(t *testing.T) {
	now := time.Now()
	items := []core.Item{
		{Name: "↑ ..", IsDir: true, IsUp: true},
		{Name: "b.txt", IsDir: false, Size: 200, MTime: now.Add(-1 * time.Hour)},
		{Name: "a.txt", IsDir: false, Size: 100, MTime: now},
		{Name: "dir", IsDir: true, Size: 0, MTime: now.Add(-2 * time.Hour)},
	}

	t.Run("SortDefault (Directories first, then name)", func(t *testing.T) {
		testItems := make([]core.Item, len(items))
		copy(testItems, items)
		SortItems(testItems, SortDefault, true)

		// Expected: .. (skipped), dir, a.txt, b.txt
		testutil.AssertEqual(t, "↑ ..", testItems[0].Name, "First item should be ..")
		testutil.AssertEqual(t, "dir", testItems[1].Name, "Second item should be dir")
		testutil.AssertEqual(t, "a.txt", testItems[2].Name, "Third item should be a.txt")
		testutil.AssertEqual(t, "b.txt", testItems[3].Name, "Fourth item should be b.txt")
	})

	t.Run("SortName", func(t *testing.T) {
		// Remove .. for easier testing of sorting all
		testItems := []core.Item{
			{Name: "b.txt"},
			{Name: "a.txt"},
			{Name: "dir"},
		}
		SortItems(testItems, SortName, false)

		testutil.AssertEqual(t, "a.txt", testItems[0].Name, "Should be a.txt")
		testutil.AssertEqual(t, "b.txt", testItems[1].Name, "Should be b.txt")
		testutil.AssertEqual(t, "dir", testItems[2].Name, "Should be dir")
	})

	t.Run("SortNameDesc", func(t *testing.T) {
		testItems := make([]core.Item, len(items))
		copy(testItems, items)
		SortItems(testItems, SortNameDesc, true)

		// Expected: .. (skipped), dir, b.txt, a.txt
		testutil.AssertEqual(t, "dir", testItems[1].Name, "Should be dir")
		testutil.AssertEqual(t, "b.txt", testItems[2].Name, "Should be b.txt")
		testutil.AssertEqual(t, "a.txt", testItems[3].Name, "Should be a.txt")
	})

	t.Run("SortSizeDesc", func(t *testing.T) {
		testItems := []core.Item{
			{Name: "b.txt", Size: 200},
			{Name: "a.txt", Size: 100},
		}
		SortItems(testItems, SortSizeDesc, false)

		testutil.AssertEqual(t, "b.txt", testItems[0].Name, "Should be b.txt (200)")
		testutil.AssertEqual(t, "a.txt", testItems[1].Name, "Should be a.txt (100)")
	})

	t.Run("SortNewest", func(t *testing.T) {
		testItems := []core.Item{
			{Name: "old", MTime: now.Add(-1 * time.Hour)},
			{Name: "new", MTime: now},
		}
		SortItems(testItems, SortNewest, false)

		testutil.AssertEqual(t, "new", testItems[0].Name, "Should be new (newest)")
	})

	t.Run("SortOldest", func(t *testing.T) {
		testItems := []core.Item{
			{Name: "old", MTime: now.Add(-1 * time.Hour)},
			{Name: "new", MTime: now},
		}
		SortItems(testItems, SortOldest, false)

		testutil.AssertEqual(t, "old", testItems[0].Name, "Should be old (oldest)")
	})

	t.Run("Empty or small list", func(t *testing.T) {
		SortItems([]core.Item{}, SortDefault, true)
		SortItems([]core.Item{{Name: ".."}}, SortDefault, true)
	})
}
