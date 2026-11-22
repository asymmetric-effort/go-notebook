// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package assets

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AcceptedImageFormats lists the MIME types accepted for images.
var AcceptedImageFormats = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".tiff": "image/tiff",
	".tif":  "image/tiff",
	".svg":  "image/svg+xml",
}

// NeedsPNGConversion returns true for formats that must be converted to PNG
// before embedding (WebP, BMP, TIFF, SVG).
func NeedsPNGConversion(ext string) bool {
	ext = strings.ToLower(ext)
	switch ext {
	case ".webp", ".bmp", ".tiff", ".tif", ".svg":
		return true
	}
	return false
}

// ResolveImage resolves an image path relative to the notebook directory,
// reads the file, and returns the MIME type and base64-encoded content.
// Returns an error if the file doesn't exist or has an unsupported format.
func ResolveImage(imgPath string, notebookDir string) (mime string, b64 string, err error) {
	// Resolve relative to notebook directory.
	fullPath := imgPath
	if !filepath.IsAbs(imgPath) {
		fullPath = filepath.Join(notebookDir, imgPath)
	}

	// Check the file exists.
	if _, err := os.Stat(fullPath); err != nil {
		return "", "", fmt.Errorf("image %q: %w", imgPath, err)
	}

	// Check the format is accepted.
	ext := strings.ToLower(filepath.Ext(fullPath))
	mimeType, ok := AcceptedImageFormats[ext]
	if !ok {
		return "", "", fmt.Errorf("image %q: unsupported format %q (accepted: png, jpg, gif, webp, bmp, tiff, svg)", imgPath, ext)
	}

	// Read and encode.
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("read image %q: %w", imgPath, err)
	}

	// For formats that need conversion, we'd convert here.
	// For now, embed as-is (PNG/JPEG/GIF are web-safe).
	if NeedsPNGConversion(ext) {
		// TODO: Convert to PNG using golang.org/x/image or Typst engine.
		// For now, embed the original format.
	}

	return mimeType, base64.StdEncoding.EncodeToString(data), nil
}

// IsAcceptedImageFormat checks if a file extension is an accepted image format.
func IsAcceptedImageFormat(ext string) bool {
	_, ok := AcceptedImageFormats[strings.ToLower(ext)]
	return ok
}
