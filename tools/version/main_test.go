// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package main

import "testing"

func TestBumpVersion(t *testing.T) {
	tests := []struct {
		current string
		bump    string
		want    string
	}{
		{"v0.0.0", "patch", "0.0.1"},
		{"v0.0.0", "minor", "0.1.0"},
		{"v0.0.0", "major", "1.0.0"},
		{"v1.2.3", "patch", "1.2.4"},
		{"v1.2.3", "minor", "1.3.0"},
		{"v1.2.3", "major", "2.0.0"},
		{"v10.20.30", "patch", "10.20.31"},
	}
	for _, tt := range tests {
		t.Run(tt.current+"_"+tt.bump, func(t *testing.T) {
			got, err := bumpVersion(tt.current, tt.bump)
			if err != nil {
				t.Fatalf("bumpVersion(%q, %q) error: %v", tt.current, tt.bump, err)
			}
			if got != tt.want {
				t.Errorf("bumpVersion(%q, %q) = %q, want %q", tt.current, tt.bump, got, tt.want)
			}
		})
	}
}

func TestBumpVersionInvalid(t *testing.T) {
	_, err := bumpVersion("v1.2.3", "invalid")
	if err == nil {
		t.Error("expected error for invalid bump")
	}

	_, err = bumpVersion("vnotaversion", "patch")
	if err == nil {
		t.Error("expected error for invalid version")
	}
}
