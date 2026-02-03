package files

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/zulfikawr/fm/internal/files/core"
)

// Analyzer handles recursive disk usage calculation
type Analyzer struct {
	fs core.FileSystem
}

// NewAnalyzer creates a new disk usage analyzer
func NewAnalyzer(fs core.FileSystem) *Analyzer {
	return &Analyzer{fs: fs}
}

// Analyze performs a recursive scan of the given path
func (a *Analyzer) Analyze(ctx context.Context, rootPath string, progress chan<- int64) (*core.AnalysisResult, error) {
	// For simplicity in this first version, we use a recursive approach.
	// We can optimize with a worker pool later if needed.
	
	result, err := a.scan(ctx, rootPath, progress)
	if err != nil {
		return nil, err
	}

	if result != nil {
		a.calculatePercentages(result)
		a.sortResult(result)
	}

	return result, nil
}

func (a *Analyzer) scan(ctx context.Context, path string, progress chan<- int64) (*core.AnalysisResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	result := &core.AnalysisResult{
		Path:        path,
		Name:        filepath.Base(path),
		IsDirectory: info.IsDir(),
	}

	if !info.IsDir() {
		result.Size = info.Size()
		if progress != nil {
			progress <- result.Size
		}
		return result, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return result, nil // Return what we have even if we can't read children
	}

	var totalSize int64
	var mu sync.Mutex

	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		
		// To avoid too many goroutines, we could limit depth or use a pool.
		// For now, let's do it synchronously for each level but we could parallelize sibling directories.
		child, err := a.scan(ctx, childPath, progress)
		if err == nil && child != nil {
			child.Parent = result
			mu.Lock()
			result.Children = append(result.Children, child)
			totalSize += child.Size
			mu.Unlock()
		}
	}

	result.Size = totalSize
	return result, nil
}

func (a *Analyzer) calculatePercentages(r *core.AnalysisResult) {
	if r.Size == 0 {
		return
	}
	for _, child := range r.Children {
		child.Percentage = float64(child.Size) / float64(r.Size)
		a.calculatePercentages(child)
	}
}

func (a *Analyzer) sortResult(r *core.AnalysisResult) {
	sort.Slice(r.Children, func(i, j int) bool {
		return r.Children[i].Size > r.Children[j].Size
	})
	for _, child := range r.Children {
		a.sortResult(child)
	}
}

// Optimized concurrent scanner
func (a *Analyzer) AnalyzeConcurrent(ctx context.Context, rootPath string, progress chan<- int64) (*core.AnalysisResult, error) {
	var totalProcessed int64
	
	// Helper to track progress
	track := func(s int64) {
		atomic.AddInt64(&totalProcessed, s)
		if progress != nil {
			select {
			case progress <- s:
			default:
			}
		}
	}

	res, err := a.scanConcurrent(ctx, rootPath, track)
	if err != nil {
		return nil, err
	}
	
	if res != nil {
		a.calculatePercentages(res)
		a.sortResult(res)
	}
	return res, nil
}

func (a *Analyzer) scanConcurrent(ctx context.Context, path string, track func(int64)) (*core.AnalysisResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	res := &core.AnalysisResult{
		Path:        path,
		Name:        filepath.Base(path),
		IsDirectory: info.IsDir(),
	}

	if !info.IsDir() {
		res.Size = info.Size()
		track(res.Size)
		return res, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return res, nil
	}

	results := make(chan *core.AnalysisResult, len(entries))
	var wg sync.WaitGroup

	for _, entry := range entries {
		wg.Add(1)
		go func(e os.DirEntry) {
			defer wg.Done()
			childPath := filepath.Join(path, e.Name())
			child, err := a.scanConcurrent(ctx, childPath, track)
			if err == nil && child != nil {
				child.Parent = res
				results <- child
			}
		}(entry)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var total int64
	for child := range results {
		res.Children = append(res.Children, child)
		total += child.Size
	}
	res.Size = total

	return res, nil
}
