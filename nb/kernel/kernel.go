// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package kernel runs notebook cells as individual processes and manages
// the cache of cell results.
package kernel

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// DefaultCellTimeout is the default time limit for cell execution.
const DefaultCellTimeout = 5 * time.Minute

// CellResult holds the outcome of running a cell.
type CellResult struct {
	Name     string
	Err      error
	ExitCode int
	Stdout   string
	Stderr   string
	Displays []DisplayEntry
	Data     []byte // Serialized exported values.
}

// DisplayEntry is a single display output from a cell.
type DisplayEntry struct {
	Mime string
	Data string
}

// CacheKey uniquely identifies a cell's cached result.
type CacheKey struct {
	SourceHash   string
	UpstreamHash string
	FileHashes   string
}

// String returns a combined hash for the cache key.
func (k CacheKey) String() string {
	h := sha256.New()
	h.Write([]byte(k.SourceHash))
	h.Write([]byte(k.UpstreamHash))
	h.Write([]byte(k.FileHashes))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Kernel manages cell execution and caching for notebook mode.
type Kernel struct {
	mu          sync.Mutex
	cache       map[string]*cacheEntry
	cellTimeout time.Duration
}

type cacheEntry struct {
	key    string
	result CellResult
}

// New creates a new kernel with the default cell timeout.
func New() *Kernel {
	return &Kernel{
		cache:       make(map[string]*cacheEntry),
		cellTimeout: DefaultCellTimeout,
	}
}

// SetCellTimeout sets the default timeout for cell execution.
func (k *Kernel) SetCellTimeout(d time.Duration) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.cellTimeout = d
}

// CellTimeout returns the current cell timeout.
func (k *Kernel) CellTimeout() time.Duration {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.cellTimeout
}

// GetCached returns the cached result for a cell, or nil if not cached
// or the cache key doesn't match (stale).
func (k *Kernel) GetCached(name string, key CacheKey) *CellResult {
	k.mu.Lock()
	defer k.mu.Unlock()
	entry, ok := k.cache[name]
	if !ok {
		return nil
	}
	if entry.key != key.String() {
		return nil
	}
	return &entry.result
}

// SetCached stores a cell result in the cache.
func (k *Kernel) SetCached(name string, key CacheKey, result CellResult) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.cache[name] = &cacheEntry{
		key:    key.String(),
		result: result,
	}
}

// Invalidate removes a cell and everything downstream from the cache.
func (k *Kernel) Invalidate(names ...string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	for _, name := range names {
		delete(k.cache, name)
	}
}

// InvalidateAll clears the entire cache.
func (k *Kernel) InvalidateAll() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.cache = make(map[string]*cacheEntry)
}

// CacheSize returns the number of cached entries.
func (k *Kernel) CacheSize() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return len(k.cache)
}

// RunCell executes a cell with the given timeout. The runner function
// is provided by the caller and handles the actual `go run` invocation.
// If the cell has a custom timeout, it overrides the kernel default.
func (k *Kernel) RunCell(ctx context.Context, name string, timeout time.Duration, runner func(ctx context.Context) (CellResult, error)) (CellResult, error) {
	if timeout == 0 {
		timeout = k.CellTimeout()
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultCh := make(chan CellResult, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := runner(ctx)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errCh:
		return CellResult{}, err
	case <-ctx.Done():
		return CellResult{}, fmt.Errorf("cell %q timed out after %v", name, timeout)
	}
}
