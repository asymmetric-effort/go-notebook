// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import "testing"

func TestChecksReturnsAll(t *testing.T) {
	c := checks()
	if len(c) != 3 {
		t.Fatalf("checks() = %d, want 3", len(c))
	}
	if c[0].name != "leakdetector" {
		t.Errorf("first check = %q, want %q", c[0].name, "leakdetector")
	}
	if c[1].name != "make lint" {
		t.Errorf("second check = %q, want %q", c[1].name, "make lint")
	}
	if c[2].name != "make test" {
		t.Errorf("third check = %q, want %q", c[2].name, "make test")
	}
}
