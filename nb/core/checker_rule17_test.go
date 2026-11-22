// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckRule17ImageExists(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "logo.png"), []byte("png"), 0o644)

	src := "![logo](logo.png)\n"
	result := Parse(src)
	errs := CheckWithOptions(result, CheckOptions{NotebookDir: dir})
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 17") {
			t.Errorf("existing image should not error: %v", e)
		}
	}
}

func TestCheckRule17ImageMissing(t *testing.T) {
	dir := t.TempDir()

	src := "![logo](missing.png)\n"
	result := Parse(src)
	errs := CheckWithOptions(result, CheckOptions{NotebookDir: dir})
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 17") && strings.Contains(e.Error(), "not found") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 17 error for missing image, got: %v", errs)
	}
}

func TestCheckRule17UnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "doc.pdf"), []byte("pdf"), 0o644)

	src := "![doc](doc.pdf)\n"
	result := Parse(src)
	errs := CheckWithOptions(result, CheckOptions{NotebookDir: dir})
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 17") && strings.Contains(e.Error(), "unsupported") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 17 error for unsupported format, got: %v", errs)
	}
}

func TestCheckRule17WebURL(t *testing.T) {
	// Web URLs should be skipped by rule 17 (handled by downloader).
	src := "![logo](https://example.com/logo.png)\n"
	result := Parse(src)
	errs := CheckWithOptions(result, CheckOptions{NotebookDir: t.TempDir()})
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 17") {
			t.Errorf("web URLs should not trigger rule 17: %v", e)
		}
	}
}

func TestCheckRule17NoNotebookDir(t *testing.T) {
	// Without NotebookDir, rule 17 is skipped.
	src := "![logo](missing.png)\n"
	result := Parse(src)
	errs := Check(result) // No options = no NotebookDir.
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 17") {
			t.Errorf("rule 17 should be skipped without NotebookDir: %v", e)
		}
	}
}

func TestCheckRule17AcceptedFormats(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.png", "b.jpg", "c.gif", "d.svg", "e.webp"} {
		os.WriteFile(filepath.Join(dir, name), []byte("data"), 0o644)
	}

	src := "![](a.png)\n![](b.jpg)\n![](c.gif)\n![](d.svg)\n![](e.webp)\n"
	result := Parse(src)
	errs := CheckWithOptions(result, CheckOptions{NotebookDir: dir})
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 17") {
			t.Errorf("accepted format should not error: %v", e)
		}
	}
}
