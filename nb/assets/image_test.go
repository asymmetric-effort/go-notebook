// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package assets

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveImagePNG(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	os.WriteFile(imgPath, []byte("fake png"), 0o644)

	mime, b64, err := ResolveImage("test.png", dir)
	if err != nil {
		t.Fatal(err)
	}
	if mime != "image/png" {
		t.Errorf("mime = %q, want image/png", mime)
	}
	decoded, _ := base64.StdEncoding.DecodeString(b64)
	if string(decoded) != "fake png" {
		t.Errorf("decoded = %q", decoded)
	}
}

func TestResolveImageJPEG(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("jpeg data"), 0o644)

	mime, _, err := ResolveImage("photo.jpg", dir)
	if err != nil {
		t.Fatal(err)
	}
	if mime != "image/jpeg" {
		t.Errorf("mime = %q", mime)
	}
}

func TestResolveImageGIF(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "anim.gif"), []byte("gif data"), 0o644)

	mime, _, err := ResolveImage("anim.gif", dir)
	if err != nil {
		t.Fatal(err)
	}
	if mime != "image/gif" {
		t.Errorf("mime = %q", mime)
	}
}

func TestResolveImageMissing(t *testing.T) {
	_, _, err := ResolveImage("missing.png", t.TempDir())
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestResolveImageUnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "doc.pdf"), []byte("pdf"), 0o644)

	_, _, err := ResolveImage("doc.pdf", dir)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestResolveImageAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "abs.png")
	os.WriteFile(imgPath, []byte("png data"), 0o644)

	mime, _, err := ResolveImage(imgPath, "/other/dir")
	if err != nil {
		t.Fatal(err)
	}
	if mime != "image/png" {
		t.Errorf("mime = %q", mime)
	}
}

func TestResolveImageWebP(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "photo.webp"), []byte("webp data"), 0o644)

	mime, _, err := ResolveImage("photo.webp", dir)
	if err != nil {
		t.Fatal(err)
	}
	if mime != "image/webp" {
		t.Errorf("mime = %q", mime)
	}
}

func TestResolveImageSVG(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "icon.svg"), []byte("<svg></svg>"), 0o644)

	mime, _, err := ResolveImage("icon.svg", dir)
	if err != nil {
		t.Fatal(err)
	}
	if mime != "image/svg+xml" {
		t.Errorf("mime = %q", mime)
	}
}

func TestNeedsPNGConversion(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".png", false},
		{".jpg", false},
		{".gif", false},
		{".webp", true},
		{".bmp", true},
		{".tiff", true},
		{".svg", true},
		{".PNG", false},
	}
	for _, tt := range tests {
		if got := NeedsPNGConversion(tt.ext); got != tt.want {
			t.Errorf("NeedsPNGConversion(%q) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}

func TestIsAcceptedImageFormat(t *testing.T) {
	if !IsAcceptedImageFormat(".png") {
		t.Error(".png should be accepted")
	}
	if !IsAcceptedImageFormat(".JPG") {
		t.Error(".JPG should be accepted (case insensitive)")
	}
	if IsAcceptedImageFormat(".pdf") {
		t.Error(".pdf should not be accepted")
	}
}

func TestOriginalFileUnchanged(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	original := []byte("original png data")
	os.WriteFile(imgPath, original, 0o644)

	_, _, err := ResolveImage("test.png", dir)
	if err != nil {
		t.Fatal(err)
	}

	// Verify original file is unchanged.
	data, _ := os.ReadFile(imgPath)
	if string(data) != string(original) {
		t.Error("original file was modified")
	}
}
