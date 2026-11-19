// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import "testing"

func TestRunReturnsZero(t *testing.T) {
	if code := run(); code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}
