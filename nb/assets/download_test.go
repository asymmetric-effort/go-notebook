// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package assets

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadAndCache(t *testing.T) {
	// Start a local HTTP server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("fake png data"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")
	dl := NewDownloader(assetsDir, 5*time.Second)

	// First download.
	path, err := dl.Download(srv.URL + "/image.png")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "fake png data" {
		t.Errorf("data = %q", data)
	}

	// Second download should use cache.
	path2, err := dl.Download(srv.URL + "/image.png")
	if err != nil {
		t.Fatal(err)
	}
	if path2 != path {
		t.Error("second download should return cached path")
	}
}

func TestDownloadNotFound(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	dir := t.TempDir()
	dl := NewDownloader(filepath.Join(dir, "assets"), 5*time.Second)

	_, err := dl.Download(srv.URL + "/missing.png")
	if err == nil {
		t.Error("expected error for 404")
	}
}

func TestDownloadTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than timeout.
		w.Write([]byte("slow"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	dl := NewDownloader(filepath.Join(dir, "assets"), 100*time.Millisecond)

	_, err := dl.Download(srv.URL + "/slow.png")
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestDownloadBadURL(t *testing.T) {
	dir := t.TempDir()
	dl := NewDownloader(filepath.Join(dir, "assets"), 5*time.Second)

	_, err := dl.Download("http://localhost:1/nonexistent")
	if err == nil {
		t.Error("expected error for bad URL")
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"http://example.com/image.png", true},
		{"https://example.com/image.png", true},
		{"HTTP://EXAMPLE.COM", true},
		{"ftp://example.com", false},
		{"image.png", false},
		{"../relative.png", false},
	}
	for _, tt := range tests {
		if got := IsURL(tt.input); got != tt.want {
			t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestURLExtension(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http://example.com/image.png", ".png"},
		{"http://example.com/image.jpg?w=100", ".jpg"},
		{"http://example.com/noext#section", ".bin"},
		{"http://example.com/file", ".bin"},
	}
	for _, tt := range tests {
		if got := urlExtension(tt.input); got != tt.want {
			t.Errorf("urlExtension(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNewDownloaderDefaultTimeout(t *testing.T) {
	dl := NewDownloader("/tmp/assets", 0)
	if dl.timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", dl.timeout, DefaultTimeout)
	}
}

func TestDownloadAssetsDir(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "deep", "nested", "assets")
	dl := NewDownloader(assetsDir, 5*time.Second)

	path, err := dl.Download(srv.URL + "/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file should exist: %v", err)
	}
}
