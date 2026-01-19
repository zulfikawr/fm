package core

// Match represents a single match within a file
type Match struct {
	Line       int    // Line number (1-based)
	Content    string // Line content
	MatchedIdx []int  // Indices of matched characters for highlighting
}

// FileResult represents all matches found within a single file
type FileResult struct {
	Path      string  // Full path to the file
	FileName  string  // Just the name for display
	Matches   []Match // List of matches in this file
	Collapsed bool    // UI state: whether this file's matches are hidden
}
