// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package assets handles downloading and caching of external resources
// (images, stylesheets) referenced in notebooks. Each URL is downloaded
// at most once into an assets directory next to the notebook.
package assets

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultTimeout is the default download timeout.
const DefaultTimeout = 30 * time.Second

// Downloader manages downloading and caching of external resources.
type Downloader struct {
	assetsDir string
	timeout   time.Duration
	client    *http.Client
}

// NewDownloader creates a downloader that caches files in the given assets directory.
func NewDownloader(assetsDir string, timeout time.Duration) *Downloader {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &Downloader{
		assetsDir: assetsDir,
		timeout:   timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Download fetches a URL and returns the local file path. If the file
// is already cached, it returns the cached path without downloading.
// The cache key is derived from the URL hash.
func (d *Downloader) Download(url string) (string, error) {
	if err := os.MkdirAll(d.assetsDir, 0o755); err != nil {
		return "", fmt.Errorf("create assets dir: %w", err)
	}

	// Compute cache key from URL.
	ext := urlExtension(url)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(url)))
	filename := hash[:16] + ext
	localPath := filepath.Join(d.assetsDir, filename)

	// Check if already cached.
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// Download.
	resp, err := d.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	f, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", localPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(localPath) // Clean up partial download.
		return "", fmt.Errorf("write %s: %w", localPath, err)
	}

	return localPath, nil
}

// IsURL checks if a reference is an HTTP or HTTPS URL.
func IsURL(ref string) bool {
	lower := strings.ToLower(ref)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// urlExtension extracts the file extension from a URL path.
func urlExtension(url string) string {
	// Strip query string and fragment.
	path := url
	if idx := strings.IndexByte(path, '?'); idx >= 0 {
		path = path[:idx]
	}
	if idx := strings.IndexByte(path, '#'); idx >= 0 {
		path = path[:idx]
	}
	ext := filepath.Ext(path)
	if ext == "" {
		return ".bin"
	}
	return ext
}
