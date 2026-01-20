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

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/files/errors"
	"github.com/zulfikawr/fm/internal/git"

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
		return nil, errors.WrapErrorWithPath(err, "Search", rootPath)
	}

	if err := g.Wait(); err != nil {
		return nil, errors.WrapErrorWithPath(err, "Search", rootPath)
	}

	return results, nil
}

func searchInFile(ctx context.Context, fs core.FileSystem, path, query string) (core.FileResult, bool) {
	fileName := fs.Base(path)
	var matches []core.Match

	// Check filename first
	if ok, matchedIdx := FuzzyMatch(fileName, query); ok {
		matches = append(matches, core.Match{
			Line:       0, // 0 indicates filename match
			Content:    fileName,
			MatchedIdx: matchedIdx,
		})
	}

	f, err := fs.Open(ctx, path)
	if err != nil {
		if len(matches) > 0 {
			return core.FileResult{
				Path:     path,
				FileName: fileName,
				Matches:  matches,
			}, true
		}
		return core.FileResult{}, false
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	if isBinary(reader) {
		if len(matches) > 0 {
			return core.FileResult{
				Path:     path,
				FileName: fileName,
				Matches:  matches,
			}, true
		}
		return core.FileResult{}, false
	}

	scanner := bufio.NewScanner(reader)
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
			FileName: fileName,
			Matches:  matches,
		}, true
	}

	return core.FileResult{}, false
}

// isBinary checks if a file is likely binary by looking for null bytes in the first 1KB.
func isBinary(r *bufio.Reader) bool {
	buf, err := r.Peek(1024)
	if err != nil && err != io.EOF && len(buf) == 0 {
		return false
	}
	return bytes.Contains(buf, []byte{0})
}

// FuzzyMatch checks if query is a substring of s (case-insensitive)
// and returns indices of matched characters.
func FuzzyMatch(s, query string) (bool, []int) {
	if query == "" {
		return true, nil
	}

	sLower := strings.ToLower(s)
	qLower := strings.ToLower(query)

	idx := strings.Index(sLower, qLower)
	if idx == -1 {
		return false, nil
	}

	// For content search, we prefer substring matches to avoid excessive noise
	matchedIdx := make([]int, len(query))
	for i := 0; i < len(query); i++ {
		matchedIdx[i] = idx + i
	}

	return true, matchedIdx
}
