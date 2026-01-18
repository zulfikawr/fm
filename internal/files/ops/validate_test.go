package ops

import (
	"context"
	"os"
	"strings"
	"testing"

	"fm/internal/testutil"
)

func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
	}{
		{"file.txt", true},
		{"", false},
		{".", false},
		{"..", false},
		{"file/with/slash", false},
		{"invalid*", false},
		{strings.Repeat("A", 300), false}, // Assuming MaxFilenameLength is around 255
	}

	for _, tt := range tests {
		err := ValidateFileName(tt.name)
		if (err == nil) != tt.isValid {
			t.Errorf("ValidateFileName(%q) error = %v, want valid = %v", tt.name, err, tt.isValid)
		}
	}
}

func TestValidateSearchQuery(t *testing.T) {
	testutil.AssertNoError(t, ValidateSearchQuery("normal search"), "Normal search should be valid")
	if err := ValidateSearchQuery("dangerous; command"); err == nil {
		t.Error("Semicolon should be invalid")
	}
}

func TestValidatePath(t *testing.T) {
	fs := testutil.NewMockFileSystem()

	t.Run("Valid path", func(t *testing.T) {
		err := ValidatePath(fs, "/base", "file.txt")
		testutil.AssertNoError(t, err, "Should be valid")
	})

	t.Run("Traversal attempt", func(t *testing.T) {
		// Mock Rel to return .. if we try to escape
		fs.RelFunc = func(basepath, targpath string) (string, error) {
			if strings.Contains(targpath, "..") {
				return "../etc/passwd", nil
			}
			return "file.txt", nil
		}
		err := ValidatePath(fs, "/base", "../etc/passwd")
		if err == nil {
			t.Error("Traversal should be invalid")
		}
	})

	t.Run("Rel error", func(t *testing.T) {
		fs.RelFunc = func(basepath, targpath string) (string, error) {
			return "", os.ErrInvalid
		}
		err := ValidatePath(fs, "/base", "file.txt")
		if err == nil {
			t.Error("Expected error when Rel fails")
		}
	})
}

func TestValidateSafePath(t *testing.T) {
	testutil.AssertNoError(t, ValidateSafePath("/some/path"), "Normal path should be valid")
	if err := ValidateSafePath(""); err == nil {
		t.Error("Empty path should be invalid")
	}
}

func TestValidateWritable(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("Writable", func(t *testing.T) {
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
			return false, nil
		}
		err := ValidateWritable(ctx, fs, "/path")
		testutil.AssertNoError(t, err, "Should be writable")
	})

	t.Run("Read-only", func(t *testing.T) {
		fs.IsReadOnlyFunc = func(ctx context.Context, path string) (bool, error) {
			return true, nil
		}
		err := ValidateWritable(ctx, fs, "/path")
		if err == nil {
			t.Error("Should return error for read-only filesystem")
		}
	})
}
