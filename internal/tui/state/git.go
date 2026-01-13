package state

// GitState holds git-related state
type GitState struct {
	Branch string // Current git branch name
	Root   string // Git repository root path
}
