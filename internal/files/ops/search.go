package ops

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"fm/internal/constants"
	"fm/internal/files/core"
	"fm/internal/git"

	"golang.org/x/sync/errgroup"
)

// Search performs a fuzzy content search within the specified directory.
func Search(ctx context.Context, fs core.FileSystem, gs git.GitService, rootPath, query string) ([]core.FileResult, error) {
	if query == "" {
		return nil, nil
	}

	// Get ignored files if git is enabled
	ignored := make(map[string]bool)
	if gs != nil && gs.IsEnabled() {
		repoRoot := gs.GetRoot(ctx, rootPath)
		if repoRoot != "" {
			cmd := exec.CommandContext(ctx, "git", "-C", repoRoot, "status", "--porcelain", "--ignored", "-uall")
			if out, err := cmd.Output(); err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "!! ") {
						path := strings.Trim(line[3:], "\" ")
						ignored[fs.Join(repoRoot, path)] = true
					}
				}
			}
		}
	}

	results := make([]core.FileResult, 0)
	var resultsMu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(constants.MaxReadDirWorkers)

	err := fs.Walk(ctx, rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Check if this specific path is ignored
		if ignored[path] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		g.Go(func() error {
			fileResult, found := searchInFile(ctx, fs, path, query)
			if found {
				resultsMu.Lock()
				results = append(results, fileResult)
				resultsMu.Unlock()
			}
			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

func searchInFile(ctx context.Context, fs core.FileSystem, path, query string) (core.FileResult, bool) {
	f, err := fs.Open(ctx, path)
	if err != nil {
		return core.FileResult{}, false
	}
	defer f.Close()

	// Handle binary check - requires ReadSeeker
	if rs, ok := f.(io.ReadSeeker); ok {
		if isBinary(rs) {
			return core.FileResult{}, false
		}
		_, _ = rs.Seek(0, 0)
	}

	var matches []core.Match
	scanner := bufio.NewScanner(f)
	lineNum := 1
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return core.FileResult{}, false
		default:
		}

		line := scanner.Text()
		if ok, matchedIdx := FuzzyMatch(line, query); ok {
			matches = append(matches, core.Match{
				Line:       lineNum,
				Content:    line,
				MatchedIdx: matchedIdx,
			})
		}
		lineNum++

		if len(matches) > 100 {
			break
		}
	}

	if len(matches) > 0 {
		return core.FileResult{
			Path:     path,
			FileName: fs.Base(path),
			Matches:  matches,
		}, true
	}

	return core.FileResult{}, false
}

// isBinary checks if a file is likely binary by looking for null bytes in the first 1KB.
func isBinary(r io.ReadSeeker) bool {
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	return bytes.Contains(buf[:n], []byte{0})
}

// FuzzyMatch checks if query is a substring of s (case-insensitive) and returns indices of matched characters.
func FuzzyMatch(s, query string) (bool, []int) {
	if query == "" {
		return true, nil
	}

	sRunes := []rune(s)
	qRunes := []rune(strings.ToLower(query))

	if len(qRunes) > len(sRunes) {
		return false, nil
	}

	for i := 0; i <= len(sRunes)-len(qRunes); i++ {
		match := true
		for j := 0; j < len(qRunes); j++ {
			if unicode.ToLower(sRunes[i+j]) != qRunes[j] {
				match = false
				break
			}
		}
		if match {
			matchedIdx := make([]int, len(qRunes))
			for j := 0; j < len(qRunes); j++ {
				matchedIdx[j] = i + j
			}
			return true, matchedIdx
		}
	}

	return false, nil
}
