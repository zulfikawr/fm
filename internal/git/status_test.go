package git

import (
	"fm/internal/testutil"
	"testing"
)

func TestParseGitStatusPorcelain(t *testing.T) {
	output := `M  internal/files/core/item.go
?? internal/files/core/item_test.go
!! internal/testutil/mock_clock.go
A  new_file.txt
D  deleted_file.txt
`
	repoRoot := "/home/user/fm"
	dirPath := "/home/user/fm"

	statuses := ParseGitStatusPorcelain(output, repoRoot, dirPath)

	t.Run("Root level parsing", func(t *testing.T) {
		testutil.AssertEqual(t, "M", statuses["internal"], "Subdirectory should have status of contained file")
		testutil.AssertEqual(t, "A", statuses["new_file.txt"], "Added file should have status A")
		testutil.AssertEqual(t, "D", statuses["deleted_file.txt"], "Deleted file should have status D")
	})

	t.Run("Subdirectory level parsing", func(t *testing.T) {
		subDirPath := "/home/user/fm/internal/files/core"
		subStatuses := ParseGitStatusPorcelain(output, repoRoot, subDirPath)

		testutil.AssertEqual(t, "M", subStatuses["item.go"], "File in subfolder should match")
		testutil.AssertEqual(t, "?", subStatuses["item_test.go"], "Untracked file should match")
	})
}
