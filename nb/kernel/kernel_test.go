// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package kernel

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewKernel(t *testing.T) {
	k := New()
	if k.CacheSize() != 0 {
		t.Errorf("CacheSize() = %d, want 0", k.CacheSize())
	}
	if k.CellTimeout() != DefaultCellTimeout {
		t.Errorf("CellTimeout() = %v, want %v", k.CellTimeout(), DefaultCellTimeout)
	}
}

func TestSetCellTimeout(t *testing.T) {
	k := New()
	k.SetCellTimeout(10 * time.Second)
	if k.CellTimeout() != 10*time.Second {
		t.Errorf("CellTimeout() = %v, want 10s", k.CellTimeout())
	}
}

func TestCacheHitAndMiss(t *testing.T) {
	k := New()
	key := CacheKey{SourceHash: "abc", UpstreamHash: "", FileHashes: ""}
	result := CellResult{Name: "test", Data: []byte("hello")}

	// Miss.
	if got := k.GetCached("test", key); got != nil {
		t.Error("expected cache miss")
	}

	// Set.
	k.SetCached("test", key, result)
	if k.CacheSize() != 1 {
		t.Errorf("CacheSize() = %d, want 1", k.CacheSize())
	}

	// Hit.
	got := k.GetCached("test", key)
	if got == nil {
		t.Fatal("expected cache hit")
	}
	if string(got.Data) != "hello" {
		t.Errorf("Data = %q, want %q", got.Data, "hello")
	}

	// Miss with different key.
	key2 := CacheKey{SourceHash: "def", UpstreamHash: "", FileHashes: ""}
	if got := k.GetCached("test", key2); got != nil {
		t.Error("expected cache miss for different key")
	}
}

func TestInvalidate(t *testing.T) {
	k := New()
	key := CacheKey{SourceHash: "abc"}
	k.SetCached("a", key, CellResult{Name: "a"})
	k.SetCached("b", key, CellResult{Name: "b"})

	k.Invalidate("a")
	if k.CacheSize() != 1 {
		t.Errorf("CacheSize() = %d, want 1", k.CacheSize())
	}
	if got := k.GetCached("a", key); got != nil {
		t.Error("expected invalidated cache miss")
	}
}

func TestInvalidateAll(t *testing.T) {
	k := New()
	key := CacheKey{SourceHash: "abc"}
	k.SetCached("a", key, CellResult{Name: "a"})
	k.SetCached("b", key, CellResult{Name: "b"})

	k.InvalidateAll()
	if k.CacheSize() != 0 {
		t.Errorf("CacheSize() = %d, want 0", k.CacheSize())
	}
}

func TestCacheKeyString(t *testing.T) {
	k1 := CacheKey{SourceHash: "a", UpstreamHash: "b", FileHashes: "c"}
	k2 := CacheKey{SourceHash: "a", UpstreamHash: "b", FileHashes: "d"}
	if k1.String() == k2.String() {
		t.Error("different keys should produce different strings")
	}
	if k1.String() != k1.String() {
		t.Error("same key should produce same string")
	}
}

func TestRunCellSuccess(t *testing.T) {
	k := New()
	result, err := k.RunCell(context.Background(), "test", time.Second, func(ctx context.Context) (CellResult, error) {
		return CellResult{Name: "test", Data: []byte("ok")}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "test" {
		t.Errorf("Name = %q", result.Name)
	}
}

func TestRunCellError(t *testing.T) {
	k := New()
	_, err := k.RunCell(context.Background(), "test", time.Second, func(ctx context.Context) (CellResult, error) {
		return CellResult{}, errors.New("cell failed")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunCellTimeout(t *testing.T) {
	k := New()
	_, err := k.RunCell(context.Background(), "test", 50*time.Millisecond, func(ctx context.Context) (CellResult, error) {
		select {
		case <-ctx.Done():
			return CellResult{}, ctx.Err()
		case <-time.After(10 * time.Second):
			return CellResult{}, nil
		}
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error = %q, want timeout", err.Error())
	}
}

func TestRunCellDefaultTimeout(t *testing.T) {
	k := New()
	k.SetCellTimeout(50 * time.Millisecond)
	_, err := k.RunCell(context.Background(), "test", 0, func(ctx context.Context) (CellResult, error) {
		select {
		case <-ctx.Done():
			return CellResult{}, ctx.Err()
		case <-time.After(10 * time.Second):
			return CellResult{}, nil
		}
	})
	if err == nil {
		t.Fatal("expected timeout with default timeout")
	}
}

func TestRunCellConcurrent(t *testing.T) {
	k := New()
	done := make(chan struct{}, 5)
	for i := 0; i < 5; i++ {
		go func() {
			k.RunCell(context.Background(), "test", time.Second, func(ctx context.Context) (CellResult, error) {
				return CellResult{Name: "test"}, nil
			})
			done <- struct{}{}
		}()
	}
	for i := 0; i < 5; i++ {
		<-done
	}
}
