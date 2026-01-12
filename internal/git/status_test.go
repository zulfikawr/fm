package git

import (
	"pgregory.net/rapid"
	"strings"
	"testing"
)

func TestParseGitStatusPorcelain_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property: Git Status Parsing
	rapid.Check(t, func(t *rapid.T) {
		// Generate random filenames (alphanumeric only, no slashes) and statuses
		filenames := rapid.SliceOfNDistinct(rapid.StringMatching("[a-zA-Z0-9]+"), 1, 10, func(s string) string { return s }).Draw(t, "filenames")
		gitStatuses := rapid.SliceOfN(rapid.SampledFrom([]string{"M ", " M", "A ", " D", "??", "!!"}), len(filenames), len(filenames)).Draw(t, "statuses")

		var builder strings.Builder
		expectedStatuses := make(map[string]string)

		for i, name := range filenames {
			status := gitStatuses[i]
			builder.WriteString(status)
			builder.WriteString(" ")
			builder.WriteString(name)
			builder.WriteString("\n")

			// Map porcelain status to single char used in FM
			char := string(status[0])
			if char == " " || char == "?" || char == "!" {
				char = string(status[1])
			}
			if status == "!!" {
				char = "!"
			}
			expectedStatuses[name] = char
		}

		parsed := ParseGitStatusPorcelain(builder.String(), "/root", "/root")

		for name, char := range expectedStatuses {
			if parsed[name] != char {
				t.Errorf("For file %s, expected status %s, got %s", name, char, parsed[name])
			}
		}
	})
}

func TestParseGitStatusPorcelain_Subdir_Property(t *testing.T) {
	// Feature: codebase-refactoring, Property: Git Status Parsing in Subdir
	rapid.Check(t, func(t *rapid.T) {
		subdir := rapid.StringMatching("[a-zA-Z0-9]+").Draw(t, "subdir")
		filenames := rapid.SliceOfNDistinct(rapid.StringMatching("[a-zA-Z0-9]+"), 1, 10, func(s string) string { return s }).Draw(t, "filenames")
		gitStatuses := rapid.SliceOfN(rapid.SampledFrom([]string{"M ", " M", "A ", " D", "??", "!!"}), len(filenames), len(filenames)).Draw(t, "statuses")

		var builder strings.Builder
		expectedStatuses := make(map[string]string)

		for i, name := range filenames {
			status := gitStatuses[i]
			builder.WriteString(status)
			builder.WriteString(" ")
			builder.WriteString(subdir)
			builder.WriteString("/")
			builder.WriteString(name)
			builder.WriteString("\n")

			char := string(status[0])
			if char == " " || char == "?" || char == "!" {
				char = string(status[1])
			}
			if status == "!!" {
				char = "!"
			}
			expectedStatuses[name] = char
		}

		// When parsing from subdir, it should extract the filenames correctly
		parsed := ParseGitStatusPorcelain(builder.String(), "/root", "/root/"+subdir)

		for name, char := range expectedStatuses {
			if parsed[name] != char {
				t.Errorf("For file %s in subdir %s, expected status %s, got %s", name, subdir, char, parsed[name])
			}
		}
	})
}
