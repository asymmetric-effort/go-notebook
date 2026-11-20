// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveImagePNG(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.png"), []byte("PNG data"), 0o644); err != nil {
		t.Fatal(err)
	}
	ref, err := ResolveImage("test.png", dir)
	if err != nil {
		t.Fatal(err)
	}
	if ref.Mime != "image/png" {
		t.Errorf("Mime = %q", ref.Mime)
	}
	if ref.NeedsPNG {
		t.Error("PNG should not need conversion")
	}
	if ref.Data == "" {
		t.Error("Data should not be empty")
	}
}

func TestResolveImageSVGNeedsConversion(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.svg"), []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	ref, err := ResolveImage("test.svg", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !ref.NeedsPNG {
		t.Error("SVG should need PNG conversion")
	}
}

func TestResolveImageAbsolutePath(t *testing.T) {
	_, err := ResolveImage("/etc/passwd", t.TempDir())
	if err == nil {
		t.Error("expected error for absolute path (rule 16)")
	}
}

func TestResolveImageParentTraversal(t *testing.T) {
	_, err := ResolveImage("../secret.png", t.TempDir())
	if err == nil {
		t.Error("expected error for ../ path (rule 16)")
	}
}

func TestResolveImageUnsupportedFormat(t *testing.T) {
	_, err := ResolveImage("file.xyz", t.TempDir())
	if err == nil {
		t.Error("expected error for unsupported format (rule 17)")
	}
}

func TestResolveImageMissing(t *testing.T) {
	_, err := ResolveImage("missing.png", t.TempDir())
	if err == nil {
		t.Error("expected error for missing file (rule 17)")
	}
}

func TestIsWebURL(t *testing.T) {
	tests := []struct {
		ref  string
		want bool
	}{
		{"https://example.com/image.png", true},
		{"http://example.com/image.png", true},
		{"HTTP://EXAMPLE.COM", true},
		{"image.png", false},
		{"./relative/path.png", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsWebURL(tt.ref); got != tt.want {
			t.Errorf("IsWebURL(%q) = %v, want %v", tt.ref, got, tt.want)
		}
	}
}
