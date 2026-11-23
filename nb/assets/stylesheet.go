// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// cssURLRegex matches url() references in CSS.
var cssURLRegex = regexp.MustCompile(`url\(\s*["']?([^"')]+)["']?\s*\)`)

// cssImportRegex matches @import "url" or @import url("url") in CSS.
var cssImportRegex = regexp.MustCompile(`@import\s+(?:url\(\s*)?["']([^"']+)["']\s*\)?\s*;`)

// EmbedStylesheet reads a CSS file and resolves its url() and @import
// references relative to the stylesheet's directory. Returns the CSS
// with all references resolved to inline data URIs or inlined content.
func EmbedStylesheet(cssPath string, notebookDir string) (string, error) {
	fullPath := cssPath
	if !filepath.IsAbs(cssPath) {
		fullPath = filepath.Join(notebookDir, cssPath)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("read stylesheet %s: %w", cssPath, err)
	}

	cssDir := filepath.Dir(fullPath)
	css := string(data)

	// Resolve @import directives by inlining the imported CSS.
	css = cssImportRegex.ReplaceAllStringFunc(css, func(match string) string {
		submatch := cssImportRegex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		importPath := submatch[1]
		if IsURL(importPath) {
			return match // Keep remote imports as-is (will be downloaded separately).
		}
		importFull := filepath.Join(cssDir, importPath)
		importData, err := os.ReadFile(importFull)
		if err != nil {
			return fmt.Sprintf("/* failed to import %s: %v */", importPath, err)
		}
		return string(importData)
	})

	// Resolve url() references to data URIs for local files.
	css = cssURLRegex.ReplaceAllStringFunc(css, func(match string) string {
		submatch := cssURLRegex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		urlPath := submatch[1]
		if IsURL(urlPath) || strings.HasPrefix(urlPath, "data:") {
			return match // Keep remote and data URLs as-is.
		}
		urlFull := filepath.Join(cssDir, urlPath)
		mime, b64, err := ResolveImage(urlFull, "")
		if err != nil {
			return match // Can't resolve — keep original.
		}
		return fmt.Sprintf("url(data:%s;base64,%s)", mime, b64)
	})

	return css, nil
}
