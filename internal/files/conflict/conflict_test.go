package conflict

import (
	"context"
	"os"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestGenerateUniqueName(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()

	t.Run("No conflict", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		name, err := GenerateUniqueName(ctx, fs, "/test/file.txt")
		testutil.AssertNoError(t, err, "GenerateUniqueName should not fail")
		testutil.AssertEqual(t, "/test/file.txt", name, "Should return original name")
	})

	t.Run("One conflict", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/test/file.txt" {
				return &testutil.MockFileInfo{FName: "file.txt"}, nil
			}
			return nil, os.ErrNotExist
		}
		name, err := GenerateUniqueName(ctx, fs, "/test/file.txt")
		testutil.AssertNoError(t, err, "GenerateUniqueName should not fail")
		testutil.AssertEqual(t, "/test/file (1).txt", name, "Should return (1) name")
	})

	t.Run("Multiple conflicts", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "/test/file.txt" || path == "/test/file (1).txt" || path == "/test/file (2).txt" {
				return &testutil.MockFileInfo{FName: "file.txt"}, nil
			}
			return nil, os.ErrNotExist
		}
		name, err := GenerateUniqueName(ctx, fs, "/test/file.txt")
		testutil.AssertNoError(t, err, "GenerateUniqueName should not fail")
		testutil.AssertEqual(t, "/test/file (3).txt", name, "Should return (3) name")
	})
}

func TestResolver(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	resolver := NewResolver()

	t.Run("No conflict returns destination", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		dst, renamed, err := resolver.Resolve(ctx, fs, "src", "dst", Ask)
		testutil.AssertNoError(t, err, "Resolve should not fail")
		testutil.AssertEqual(t, "dst", dst, "Should return dst")
		testutil.AssertEqual(t, false, renamed, "Should not be renamed")
	})

	t.Run("Conflict Ask returns error", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return &testutil.MockFileInfo{FName: "dst"}, nil
		}
		_, _, err := resolver.Resolve(ctx, fs, "src", "dst", Ask)
		if err == nil {
			t.Fatal("expected conflict error")
		}
		var target *ConflictError
		testutil.AssertErrorType(t, err, &target, "Should be ConflictError")
	})

	t.Run("Conflict Overwrite returns destination", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return &testutil.MockFileInfo{FName: "dst"}, nil
		}
		dst, renamed, err := resolver.Resolve(ctx, fs, "src", "dst", Overwrite)
		testutil.AssertNoError(t, err, "Resolve should not fail")
		testutil.AssertEqual(t, "dst", dst, "Should return dst")
		testutil.AssertEqual(t, false, renamed, "Should not be renamed")
	})

	t.Run("Conflict Skip returns empty", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			return &testutil.MockFileInfo{FName: "dst"}, nil
		}
		dst, renamed, err := resolver.Resolve(ctx, fs, "src", "dst", Skip)
		testutil.AssertNoError(t, err, "Resolve should not fail")
		testutil.AssertEqual(t, "", dst, "Should return empty string")
		testutil.AssertEqual(t, false, renamed, "Should not be renamed")
	})

	t.Run("Conflict Rename returns new name", func(t *testing.T) {
		fs.StatFunc = func(ctx context.Context, path string) (os.FileInfo, error) {
			if path == "dst.txt" {
				return &testutil.MockFileInfo{FName: "dst.txt"}, nil
			}
			return nil, os.ErrNotExist
		}
		fs.ExtFunc = func(path string) string { return ".txt" }
		dst, renamed, err := resolver.Resolve(ctx, fs, "src.txt", "dst.txt", Rename)
		testutil.AssertNoError(t, err, "Resolve should not fail")
		testutil.AssertEqual(t, "dst (1).txt", dst, "Should return renamed path")
		testutil.AssertEqual(t, true, renamed, "Should be renamed")
	})
}

func TestValidateSecurePath(t *testing.T) {
	fs := testutil.NewMockFileSystem()

	fs.AbsFunc = func(path string) (string, error) {
		if path == "/base/../outside.txt" {
			return "/outside.txt", nil
		}
		return path, nil
	}
	fs.SeparatorFunc = func() string {
		return "/"
	}
	fs.JoinFunc = func(elem ...string) string {
		if len(elem) == 2 && elem[0] == "/base" && elem[1] == "../outside.txt" {
			return "/base/../outside.txt"
		}
		if len(elem) == 2 && elem[0] == "/base" && elem[1] == "inside.txt" {
			return "/base/inside.txt"
		}
		return ""
	}
	fs.RelFunc = func(base, target string) (string, error) {
		if base == "/base" && target == "/outside.txt" {
			return "../outside.txt", nil
		}
		if base == "/base" && target == "/base/inside.txt" {
			return "inside.txt", nil
		}
		return "", nil
	}

	t.Run("Inside path", func(t *testing.T) {
		path, err := ValidateSecurePath(fs, "/base", "inside.txt")
		testutil.AssertNoError(t, err, "Should be valid")
		testutil.AssertEqual(t, "/base/inside.txt", path, "Path should match")
	})

	t.Run("Outside path", func(t *testing.T) {
		_, err := ValidateSecurePath(fs, "/base", "../outside.txt")
		if err == nil {
			t.Fatal("Expected error for outside path")
		}
	})
}
