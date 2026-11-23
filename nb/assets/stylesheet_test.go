// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbedStylesheetSimple(t *testing.T) {
	dir := t.TempDir()
	css := "body { color: red; }"
	os.WriteFile(filepath.Join(dir, "style.css"), []byte(css), 0o644)

	result, err := EmbedStylesheet("style.css", dir)
	if err != nil {
		t.Fatal(err)
	}
	if result != css {
		t.Errorf("simple CSS should pass through: %q", result)
	}
}

func TestEmbedStylesheetImport(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "base.css"), []byte("h1 { font-size: 2em; }"), 0o644)
	os.WriteFile(filepath.Join(dir, "main.css"), []byte("@import \"base.css\";\nbody { margin: 0; }"), 0o644)

	result, err := EmbedStylesheet("main.css", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "font-size: 2em") {
		t.Error("imported CSS should be inlined")
	}
	if !strings.Contains(result, "margin: 0") {
		t.Error("original CSS should be preserved")
	}
}

func TestEmbedStylesheetMissingImport(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.css"), []byte("@import \"missing.css\";"), 0o644)

	result, err := EmbedStylesheet("main.css", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "failed to import") {
		t.Error("missing import should produce comment")
	}
}

func TestEmbedStylesheetMissingFile(t *testing.T) {
	_, err := EmbedStylesheet("/nonexistent/style.css", "")
	if err == nil {
		t.Error("expected error for missing stylesheet")
	}
}

func TestEmbedStylesheetRemoteURL(t *testing.T) {
	dir := t.TempDir()
	css := "@import \"https://example.com/base.css\";\nbody { color: blue; }"
	os.WriteFile(filepath.Join(dir, "style.css"), []byte(css), 0o644)

	result, err := EmbedStylesheet("style.css", dir)
	if err != nil {
		t.Fatal(err)
	}
	// Remote imports should be kept as-is.
	if !strings.Contains(result, "https://example.com/base.css") {
		t.Error("remote import should be preserved")
	}
}
