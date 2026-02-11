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
		FSize:    100,
		FMode:    0644,
		FModTime: now,
		FIsDir:   false,
	}

	item := NewItem(info, "/path/test.txt", "M")

	if item.Name != "test.txt" {
		t.Errorf("Expected test.txt, got %s", item.Name)
	}
	if item.Metadata.Size != 100 {
		t.Errorf("Expected 100, got %d", item.Metadata.Size)
	}
	if item.Display.GitStatus != "M" {
		t.Errorf("Expected M, got %s", item.Display.GitStatus)
	}
	if !item.State.HasMetadata {
		t.Error("Expected HasMetadata to be true")
	}
}

func TestItem_Formatting(t *testing.T) {
	item := Item{
		Metadata: ItemMetadata{
			Size:  1024,
			MTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	// 1024 bytes -> 1.0 KB (standard format)
	item.UpdateFormatting(1, 0)

	if item.Display.FormattedSize != "1.0 KB" {
		t.Errorf("Expected 1.0 KB, got %s", item.Display.FormattedSize)
	}
}

func TestItem_IsArchive(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"test.zip", true},
		{"test.tar", true},
		{"test.tar.gz", true},
		{"test.txt", false},
		{"test", false},
	}

	for i := range tests {
		tt := tests[i]
		item := Item{Name: tt.name}
		if item.IsArchive() != tt.expected {
			t.Errorf("IsArchive(%s) = %v, expected %v", tt.name, item.IsArchive(), tt.expected)
		}
	}
}
