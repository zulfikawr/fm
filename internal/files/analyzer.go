package files

import (
	"context"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/zulfikawr/fm/internal/files/core"
	"github.com/zulfikawr/fm/internal/logger"
	"golang.org/x/sync/semaphore"
)

// Analyzer handles recursive disk usage calculation
type Analyzer struct {
	fs core.FileSystem
}

// NewAnalyzer creates a new disk usage analyzer
func NewAnalyzer(fs core.FileSystem) *Analyzer {
	return &Analyzer{fs: fs}
}

// AnalyzeConcurrent performs a recursive scan of the given path using a worker pool
func (a *Analyzer) AnalyzeConcurrent(ctx context.Context, rootPath string, progress chan<- int64) (*core.AnalysisResult, error) {
	// Semaphore limits concurrent IO operations (ReadDir/Stat)
	sem := semaphore.NewWeighted(64)
	var totalProcessed int64

	track := func(s int64) {
		atomic.AddInt64(&totalProcessed, s)
		if progress != nil {
			select {
			case progress <- s:
			default:
			}
		}
	}

	// Capture root device ID to implement "One Filesystem" rule
	var rootDev uint64
	info, err := a.fs.Lstat(ctx, rootPath)
	if err == nil {
		rootDev = GetDeviceID(info)
	}

	res, err := a.scanConcurrent(ctx, rootPath, track, sem, rootDev)
	if err != nil {
		return nil, err
	}

	if res != nil {
		a.calculatePercentages(res)
		a.sortResult(res)
	}
	return res, nil
}

func (a *Analyzer) scanConcurrent(ctx context.Context, path string, track func(int64), sem *semaphore.Weighted, rootDev uint64) (*core.AnalysisResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 1. Get file info (Throttled via FileSystem interface)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	info, err := a.fs.Lstat(ctx, path)
	sem.Release(1)

	if err != nil {
		return nil, nil
	}

	// Implement One Filesystem Rule:
	// If this is a directory and its device ID differs from root, skip recursion
	currentDev := GetDeviceID(info)
	if rootDev != 0 && currentDev != 0 && currentDev != rootDev {
		return &core.AnalysisResult{
			Path:        path,
			Name:        a.fs.Base(path) + " [Mount Point]",
			IsDirectory: true,
			Size:        0,
		}, nil
	}

	res := &core.AnalysisResult{
		Path:        path,
		Name:        a.fs.Base(path),
		IsDirectory: info.IsDir(),
	}

	// 2. Handle non-directories or symlinks
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		res.Size = info.Size()
		track(res.Size)
		return res, nil
	}

	// 3. Read directory entries (Throttled via FileSystem interface)
	if err := sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	entries, err := a.fs.ReadDirEntries(ctx, path)
	sem.Release(1)

	if err != nil {
		return res, nil
	}

	if len(entries) == 0 {
		return res, nil
	}

	// 4. Process children concurrently
	results := make(chan *core.AnalysisResult, len(entries))
	var wg sync.WaitGroup

	for i := range entries {
		// Check context in loop
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		default:
		}
		
		entry := entries[i]
		wg.Add(1)
		childPath := a.fs.Join(path, entry.Name())

		go func(p string) {
			defer wg.Done()
			child, err := a.scanConcurrent(ctx, p, track, sem, rootDev)
			if err != nil {
				logger.LogIfError(err, "Recursive scan failed for "+p)
			}
			if child != nil {
				results <- child
			}
		}(childPath)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var total int64
	for child := range results {
		child.Parent = res
		res.Children = append(res.Children, child)
		total += child.Size
	}
	res.Size = total

	return res, nil
}

func (a *Analyzer) calculatePercentages(r *core.AnalysisResult) {
	if r.Size == 0 {
		return
	}
	for i := range r.Children {
		child := r.Children[i]
		child.Percentage = float64(child.Size) / float64(r.Size)
		a.calculatePercentages(child)
	}
}

func (a *Analyzer) sortResult(r *core.AnalysisResult) {
	sort.Slice(r.Children, func(i, j int) bool {
		return r.Children[i].Size > r.Children[j].Size
	})
	for i := range r.Children {
		a.sortResult(r.Children[i])
	}
}
