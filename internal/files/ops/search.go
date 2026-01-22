package ops

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zulfikawr/fm/internal/constants"
	"github.com/zulfikawr/fm/internal/files/core"

	"golang.org/x/sync/errgroup"
)

// Search performs a fuzzy content search within the specified directory.
func Search(opts SearchOptions) ([]core.FileResult, error) {
	if opts.Query == "" {
		return nil, nil
	}

	// Get ignored files if git is enabled
	ignored := make(map[string]bool)
	if opts.Git != nil && opts.Git.IsEnabled() {
		gitIgnored, err := opts.Git.GetIgnoredFiles(opts.OpCtx.Context, opts.Root)
		if err == nil {
			for _, p := range gitIgnored {
				ignored[p] = true
			}
		}
	}

	var results []core.FileResult
	var mu sync.Mutex

	g, ctx := errgroup.WithContext(opts.OpCtx.Context)
	g.SetLimit(constants.MaxSearchWorkers)

	err := opts.OpCtx.FS.Walk(ctx, opts.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			if ignored[path] {
				return filepath.SkipDir
			}
			return nil
		}

		if ignored[path] {
			return nil
		}

		// Parallel search within files
		fpath := path // Capture for closure
		g.Go(func() error {
			// Create a copy of options for each worker with the specific file path
			fileOpts := opts
			fileOpts.Root = fpath
			res, found := searchInFile(fileOpts)
			if found {
				mu.Lock()
				results = append(results, res)
				mu.Unlock()
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

func searchInFile(opts SearchOptions) (core.FileResult, bool) {
	path := opts.Root
	query := opts.Query
	fs := opts.OpCtx.FS
	ctx := opts.OpCtx.Context

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
