// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import "testing"

func TestChecksReturnsAll(t *testing.T) {
	c := checks()
	if len(c) != 4 {
		t.Fatalf("checks() = %d, want 4", len(c))
	}
	expected := []string{"leakdetector", "govulncheck", "make lint", "make test"}
	for i, name := range expected {
		if c[i].name != name {
			t.Errorf("check[%d] = %q, want %q", i, c[i].name, name)
		}
	}
}
