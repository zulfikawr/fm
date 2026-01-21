package core

import (
	"testing"
	"time"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestNewItem(t *testing.T) {
	now := time.Now()
	info := &testutil.MockFileInfo{
		FName:    "test.txt",
		FSize:    1234,
		FMode:    0644,
		FModTime: now,
		FIsDir:   false,
	}

	item := NewItem(info, "/path/test.txt", "M")

	testutil.AssertEqual(t, "test.txt", item.Name, "Name should match")
	testutil.AssertEqual(t, int64(1234), item.Size, "Size should match")
	testutil.AssertEqual(t, "M", item.GitStatus, "Git status should match")
	testutil.AssertEqual(t, false, item.IsDir, "IsDir should be false")
	testutil.AssertEqual(t, now, item.MTime, "ModTime should match")
}

func TestUpdateFormatting(t *testing.T) {
	item := Item{
		Size:  1024 * 1024,
		MTime: time.Date(2026, 1, 14, 12, 0, 0, 0, time.UTC),
	}

	// Format indices: Size 1 (Full), Date 0 (Default "02/01/2006 15:04")
	item.UpdateFormatting(1, 0)

	testutil.AssertEqual(t, "1.0 MB", item.FormattedSize, "Formatted size should be 1.0 MB")
	testutil.AssertEqual(t, "14/01/2026 12:00", item.FormattedDate, "Formatted date should match")
}

func TestUpdateFormattingDeleted(t *testing.T) {
	t.Run("Deleted file hides metadata", func(t *testing.T) {
		item := Item{
			Name:      "deleted.txt",
			GitStatus: "D",
			IsDir:     false,
			Size:      0,
			MTime:     time.Time{},
		}
		item.UpdateFormatting(1, 0)
		testutil.AssertEqual(t, "", item.FormattedSize, "Formatted size should be empty")
		testutil.AssertEqual(t, "", item.FormattedDate, "Formatted date should be empty")
	})

	t.Run("Deleted directory keeps date", func(t *testing.T) {
		now := time.Date(2026, 1, 14, 12, 0, 0, 0, time.UTC)
		item := Item{
			Name:      "deleted_dir",
			GitStatus: "D",
			IsDir:     true,
			Size:      -1,
			MTime:     now,
		}
		item.UpdateFormatting(1, 0)
		testutil.AssertEqual(t, "", item.FormattedSize, "Formatted size should be empty for dir")
		testutil.AssertEqual(t, "14/01/2026 12:00", item.FormattedDate, "Formatted date should match")
	})
}
