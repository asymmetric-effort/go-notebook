// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AcceptedImageFormats lists the image formats the notebook accepts.
var AcceptedImageFormats = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".tiff": "image/tiff",
	".svg":  "image/svg+xml",
}

// EmbeddableFormats are the formats embedded directly (no conversion needed).
var EmbeddableFormats = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
}

// ImageRef represents a resolved image reference.
type ImageRef struct {
	Path     string // Original path from Markdown.
	Mime     string // Resolved MIME type.
	Data     string // Base64-encoded data for embedding.
	NeedsPNG bool   // True if the format needs conversion to PNG.
}

// ResolveImage reads an image file relative to the notebook directory and
// returns its reference with base64-encoded data for embedding.
func ResolveImage(imgPath, notebookDir string) (ImageRef, error) {
	// Check the path stays inside the notebook directory (rule 16).
	if filepath.IsAbs(imgPath) {
		return ImageRef{}, fmt.Errorf("image path %q is absolute (rule 16)", imgPath)
	}
	if strings.Contains(imgPath, "..") {
		return ImageRef{}, fmt.Errorf("image path %q leaves notebook directory (rule 16)", imgPath)
	}

	ext := strings.ToLower(filepath.Ext(imgPath))
	mime, ok := AcceptedImageFormats[ext]
	if !ok {
		return ImageRef{}, fmt.Errorf("image %q has unsupported format %q (rule 17)", imgPath, ext)
	}

	fullPath := filepath.Join(notebookDir, imgPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return ImageRef{}, fmt.Errorf("image %q: %w (rule 17)", imgPath, err)
	}

	ref := ImageRef{
		Path:     imgPath,
		Mime:     mime,
		Data:     base64.StdEncoding.EncodeToString(data),
		NeedsPNG: !EmbeddableFormats[ext],
	}

	return ref, nil
}

// IsWebURL checks if an image reference is a URL.
func IsWebURL(ref string) bool {
	lower := strings.ToLower(ref)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}
