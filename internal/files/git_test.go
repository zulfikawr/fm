package files

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetGitStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fm-git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Helper to run git commands
	git := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", tmpDir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(out))
		}
	}

	// Initialize git repo
	git("init")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test User")

	// Create a file and commit it
	os.WriteFile(filepath.Join(tmpDir, "committed.txt"), []byte("committed"), 0644)
	git("add", "committed.txt")
	git("commit", "-m", "initial commit")

	// Create a modified file
	os.WriteFile(filepath.Join(tmpDir, "committed.txt"), []byte("modified"), 0644)

	// Create an untracked file
	os.WriteFile(filepath.Join(tmpDir, "untracked.txt"), []byte("untracked"), 0644)

	// Create a deleted file (ghost)
	os.WriteFile(filepath.Join(tmpDir, "deleted.txt"), []byte("to be deleted"), 0644)
	git("add", "deleted.txt")
	git("commit", "-m", "add file to delete")
	os.Remove(filepath.Join(tmpDir, "deleted.txt"))

	// Test Renamed file setup (MUST commit before staging other things)
	os.WriteFile(filepath.Join(tmpDir, "oldname.txt"), []byte("rename me"), 0644)
	git("add", "oldname.txt")
	git("commit", "-m", "pre-rename")
	git("mv", "oldname.txt", "newname.txt")

	// Create a staged file
	os.WriteFile(filepath.Join(tmpDir, "staged.txt"), []byte("staged"), 0644)
	git("add", "staged.txt")

	// Test Quoted path (space in filename)
	os.WriteFile(filepath.Join(tmpDir, "file with space.txt"), []byte("space"), 0644)

	// Test Ignored file
	os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("ignored.txt\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "ignored.txt"), []byte("ignore me"), 0644)

	statuses, branch := GetGitStatus(tmpDir)

	if branch == "" {
		t.Error("Expected branch name, got empty string")
	}

	if statuses["deleted.txt"] != "D" {
		t.Errorf("Expected status D for deleted.txt, got %s", statuses["deleted.txt"])
	}

	if statuses["ignored.txt"] != "!" {
		t.Errorf("Expected status ! for ignored.txt, got %s", statuses["ignored.txt"])
	}
}
