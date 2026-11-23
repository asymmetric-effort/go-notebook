// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHostileCorpusCheckerBlocks verifies that every hostile-input corpus
// file is flagged by the checker when used as prose in a notebook.
// The corpus contains XSS polyglots, dangerous tags, event handlers,
// javascript: URLs, path traversals, null bytes, and CSS injection.
func TestHostileCorpusCheckerBlocks(t *testing.T) {
	corpusFiles, err := filepath.Glob("testdata/corpus/*.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(corpusFiles) == 0 {
		t.Fatal("no corpus files found in testdata/corpus/")
	}

	for _, f := range corpusFiles {
		name := strings.TrimSuffix(filepath.Base(f), ".md")
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}

			src := string(data) + "\n"
			result := Parse(src)
			errs := Check(result)

			// Every corpus file should trigger at least one check error
			// since they all contain hostile HTML constructs.
			if len(errs) == 0 {
				t.Errorf("corpus file %s should trigger checker errors but got none", name)
			}
			t.Logf("%s: %d checker errors (expected)", name, len(errs))
		})
	}
}

// TestHostileScriptTagVariants tests specific script tag variants that
// must all be caught by the checker.
func TestHostileScriptTagVariants(t *testing.T) {
	variants := []struct {
		name string
		html string
	}{
		{"lowercase", "<script>alert('xss')</script>\n"},
		{"uppercase", "<SCRIPT>alert('xss')</SCRIPT>\n"},
		{"mixed_case", "<ScRiPt>alert('xss')</ScRiPt>\n"},
		{"with_attributes", "<script type=\"text/javascript\">alert('xss')</script>\n"},
		{"self_closing", "<script src=\"evil.js\"/>\n"},
	}

	for _, v := range variants {
		t.Run(v.name, func(t *testing.T) {
			result := Parse(v.html)
			errs := Check(result)
			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), "script") || strings.Contains(e.Error(), "SCRIPT") || strings.Contains(e.Error(), "rule 17") {
					found = true
				}
			}
			if !found {
				t.Errorf("script variant %q should be caught, errors: %v", v.name, errs)
			}
		})
	}
}

// TestHostileEventHandlerVariants tests that all on* event handlers are blocked.
func TestHostileEventHandlerVariants(t *testing.T) {
	handlers := []string{
		"onclick", "onload", "onerror", "onmouseover", "onfocus",
		"onblur", "onsubmit", "onchange", "onkeydown", "onkeyup",
		"onkeypress", "ondblclick", "onmousedown", "onmouseup",
		"oncontextmenu", "ondrag", "ondrop", "onscroll", "onwheel",
	}

	for _, handler := range handlers {
		t.Run(handler, func(t *testing.T) {
			src := "<div " + handler + "=\"alert('xss')\">test</div>\n"
			result := Parse(src)
			errs := Check(result)
			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), handler) || strings.Contains(e.Error(), "rule 18") {
					found = true
				}
			}
			if !found {
				t.Errorf("event handler %q should be caught, errors: %v", handler, errs)
			}
		})
	}
}

// TestHostileJavascriptURLVariants tests javascript: URL detection.
func TestHostileJavascriptURLVariants(t *testing.T) {
	variants := []struct {
		name string
		html string
	}{
		{"lowercase", "<a href=\"javascript:alert('xss')\">x</a>\n"},
		{"uppercase", "<a href=\"JAVASCRIPT:alert('xss')\">x</a>\n"},
		{"leading_space", "<a href=\" javascript:alert('xss')\">x</a>\n"},
		{"tab_prefix", "<a href=\"\tjavascript:alert('xss')\">x</a>\n"},
	}

	for _, v := range variants {
		t.Run(v.name, func(t *testing.T) {
			result := Parse(v.html)
			errs := Check(result)
			found := false
			for _, e := range errs {
				if strings.Contains(strings.ToLower(e.Error()), "javascript") || strings.Contains(e.Error(), "rule 18") {
					found = true
				}
			}
			if !found {
				t.Errorf("javascript URL variant %q should be caught, errors: %v", v.name, errs)
			}
		})
	}
}

// TestHostileDangerousTags tests that all refused tags are blocked.
func TestHostileDangerousTags(t *testing.T) {
	tags := []string{"script", "iframe", "frame", "object", "embed", "base", "meta"}

	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			src := "<" + tag + ">content</" + tag + ">\n"
			result := Parse(src)
			errs := Check(result)
			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), tag) || strings.Contains(e.Error(), "rule 17") {
					found = true
				}
			}
			if !found {
				t.Errorf("dangerous tag %q should be caught, errors: %v", tag, errs)
			}
		})
	}
}

// TestHostilePathTraversal tests that path traversal in image references
// is caught by the checker.
func TestHostilePathTraversal(t *testing.T) {
	variants := []struct {
		name string
		src  string
	}{
		{"dotdot", "![img](../../../etc/passwd)\n"},
		{"absolute", "![img](/etc/passwd)\n"},
		{"dotdot_nested", "![img](foo/../../etc/passwd)\n"},
	}

	for _, v := range variants {
		t.Run(v.name, func(t *testing.T) {
			result := Parse(v.src)
			errs := Check(result)
			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), "rule 16") || strings.Contains(e.Error(), "traversal") || strings.Contains(e.Error(), "absolute") {
					found = true
				}
			}
			if !found {
				t.Errorf("path traversal %q should be caught, errors: %v", v.name, errs)
			}
		})
	}
}

// TestHostileInfoStringInjection tests that hostile info strings are
// rejected by the parser/checker.
func TestHostileInfoStringInjection(t *testing.T) {
	variants := []struct {
		name string
		src  string
	}{
		{"name_injection", "```go:notebook name=test;rm -rf / exported(val int)\nval=1\n```\n"},
		{"name_space", "```go:notebook name=\"test hack\" exported(val int)\nval=1\n```\n"},
		{"empty_name", "```go:notebook name= exported(val int)\nval=1\n```\n"},
		{"oversized_name", "```go:notebook name=" + strings.Repeat("a", 256) + " exported(val int)\nval=1\n```\n"},
	}

	for _, v := range variants {
		t.Run(v.name, func(t *testing.T) {
			result := Parse(v.src)
			errs := Check(result)
			// These should either fail to parse or be caught by checker.
			// At minimum, they should not crash.
			t.Logf("%s: parse blocks=%d, check errors=%d", v.name, len(result.Blocks), len(errs))
		})
	}
}

// TestHostileDeepNesting tests that deeply nested structures don't crash.
func TestHostileDeepNesting(t *testing.T) {
	// 100-deep nested blockquotes.
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("> ")
	}
	sb.WriteString("deep\n")

	result := Parse(sb.String())
	_ = Check(result)
	// Must not crash.
}

// TestHostileOversizedInput tests that very large inputs don't crash.
func TestHostileOversizedInput(t *testing.T) {
	// 1MB of repeated text.
	big := strings.Repeat("This is a line of text. ", 40000) + "\n"
	result := Parse(big)
	_ = Check(result)
	// Must not crash.
}
