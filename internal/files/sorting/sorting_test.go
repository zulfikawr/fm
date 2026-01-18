package sorting

import (
	"testing"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/testutil"
)

func TestSortItems(t *testing.T) {
	items := []core.Item{
		{Name: "↑ ..", IsDir: true, IsUp: true},
		{Name: "b.txt", IsDir: false, Size: 200},
		{Name: "a.txt", IsDir: false, Size: 100},
		{Name: "dir", IsDir: true, Size: 0},
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

	t.Run("SortNameDesc", func(t *testing.T) {
		testItems := make([]core.Item, len(items))
		copy(testItems, items)
		SortItems(testItems, SortNameDesc, true)

		// Expected: .. (skipped), dir, b.txt, a.txt
		testutil.AssertEqual(t, "dir", testItems[1].Name, "Should be dir")
		testutil.AssertEqual(t, "b.txt", testItems[2].Name, "Should be b.txt")
		testutil.AssertEqual(t, "a.txt", testItems[3].Name, "Should be a.txt")
	})

	t.Run("SortSizeAsc", func(t *testing.T) {
		testItems := make([]core.Item, len(items))
		copy(testItems, items)
		SortItems(testItems, SortSizeAsc, true)

		// dir (0) < a (100) < b (200)
		testutil.AssertEqual(t, "dir", testItems[1].Name, "Should be dir")
		testutil.AssertEqual(t, "a.txt", testItems[2].Name, "Should be a.txt")
		testutil.AssertEqual(t, "b.txt", testItems[3].Name, "Should be b.txt")
	})
}
