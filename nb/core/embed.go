// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractEmbedPatterns parses //go:embed directives from Go source code
// and returns the list of patterns. Each directive can have multiple
// space-separated patterns on the same line.
func ExtractEmbedPatterns(src string) []string {
	var patterns []string
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "//go:embed ") {
			continue
		}
		// Everything after "//go:embed " is space-separated patterns.
		rest := strings.TrimPrefix(trimmed, "//go:embed ")
		for _, p := range strings.Fields(rest) {
			if p != "" {
				patterns = append(patterns, p)
			}
		}
	}
	return patterns
}

// CopyEmbedFiles copies files referenced by //go:embed directives from
// the notebook's directory into the generated package directory. This is
// needed because embed patterns resolve relative to the package directory,
// and the generated package is in a different location.
func CopyEmbedFiles(patterns []string, srcDir, dstDir string) error {
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(srcDir, pattern))
		if err != nil {
			return fmt.Errorf("embed pattern %q: %w", pattern, err)
		}
		for _, src := range matches {
			rel, err := filepath.Rel(srcDir, src)
			if err != nil {
				return fmt.Errorf("embed rel path %q: %w", src, err)
			}
			dst := filepath.Join(dstDir, rel)
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("embed copy %q: %w", rel, err)
			}
		}
	}
	return nil
}

// copyFile copies a single file from src to dst, creating directories as needed.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
