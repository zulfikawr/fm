package ops

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zulfikawr/fm/internal/testutil"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		s        string
		query    string
		expected bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "world", true},
		{"Hello World", "word", false}, // substring match: no
		{"Hello World", "lo wo", true},
		{"Hello World", "wrong", false},
		{"", "test", false},
		{"anything", "", true},
	}
	for _, tt := range tests {
		result, _ := FuzzyMatch(tt.s, tt.query)
		if result != tt.expected {
			t.Errorf("FuzzyMatch(%q, %q) = %v, want %v", tt.s, tt.query, result, tt.expected)
		}
	}
}

func TestIsBinary(t *testing.T) {
	t.Run("Text file", func(t *testing.T) {
		data := []byte("this is some text")
		r := testutil.NewMockFile("test.txt", data)
		reader := bufio.NewReader(r)
		if isBinary(reader) {
			t.Error("Expected text file to not be identified as binary")
		}
	})
	t.Run("Binary file", func(t *testing.T) {
		data := []byte{0x00, 0x01, 0x02}
		r := testutil.NewMockFile("test.bin", data)
		reader := bufio.NewReader(r)
		if !isBinary(reader) {
			t.Error("Expected binary file to be identified as binary")
		}
	})
}

func TestSearch_MatchLimit(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	content := strings.Repeat("match\n", 101)
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile("large.txt", []byte(content)), nil
	}
	result, found := searchInFile(ctx, fs, "/large.txt", "match")
	if !found {
		t.Fatal("Should have found matches")
	}
	if len(result.Matches) != 101 {
		t.Errorf("Expected 101 matches, got %d", len(result.Matches))
	}
}

type mockReadCloser struct{ io.Reader }

func (m *mockReadCloser) Close() error { return nil }

func TestSearch(t *testing.T) {
	ctx := context.Background()
	fs := testutil.NewMockFileSystem()
	fs.WalkFunc = func(ctx context.Context, root string, walkFn func(string, os.FileInfo, error) error) error {
		info := &testutil.MockFileInfo{NameStr: "test.txt", IsDirBool: false}
		return walkFn("/test.txt", info, nil)
	}
	fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
		return testutil.NewMockFile(fs.Base(path), []byte("match this query\nsecond line")), nil
	}
	fs.BaseFunc = func(path string) string {
		return "test.txt"
	}

	t.Run("Empty Query", func(t *testing.T) {
		results, err := Search(ctx, fs, nil, "/", "")
		testutil.AssertNoError(t, err, "Search should not error on empty query")
		if len(results) != 0 {
			t.Errorf("Expected 0 results for empty query, got %d", len(results))
		}
	})

	t.Run("Successful Match", func(t *testing.T) {
		results, err := Search(ctx, fs, nil, "/", "match")
		testutil.AssertNoError(t, err, "Search should succeed")
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	t.Run("Skip .git directory", func(t *testing.T) {
		fs.WalkFunc = func(ctx context.Context, root string, walkFn func(string, os.FileInfo, error) error) error {
			info := &testutil.MockFileInfo{NameStr: ".git", IsDirBool: true}
			err := walkFn("/.git", info, nil)
			if err == filepath.SkipDir {
				return nil
			}
			return err
		}
		results, err := Search(ctx, fs, nil, "/", "any")
		testutil.AssertNoError(t, err, "Search should succeed")
		if len(results) != 0 {
			t.Errorf("Expected 0 results from .git directory, got %d", len(results))
		}
	})

	t.Run("Non-seeker file", func(t *testing.T) {
		fs.WalkFunc = func(ctx context.Context, root string, walkFn func(string, os.FileInfo, error) error) error {
			info := &testutil.MockFileInfo{NameStr: "nonseeker.txt", IsDirBool: false}
			return walkFn("/nonseeker.txt", info, nil)
		}
		fs.OpenFunc = func(ctx context.Context, path string) (io.ReadCloser, error) {
			return &mockReadCloser{Reader: strings.NewReader("match")}, nil
		}
		results, err := Search(ctx, fs, nil, "/", "match")
		testutil.AssertNoError(t, err, "Search should succeed even if file is not seeker")
		if len(results) == 0 {
			t.Error("Expected results even if file is not seeker")
		}
	})

	t.Run("Walk error", func(t *testing.T) {
		fs.WalkFunc = func(ctx context.Context, root string, walkFn func(string, os.FileInfo, error) error) error {
			return os.ErrPermission
		}
		_, err := Search(ctx, fs, nil, "/", "match")
		if err == nil {
			t.Error("Expected error when Walk fails")
		}
	})
}
