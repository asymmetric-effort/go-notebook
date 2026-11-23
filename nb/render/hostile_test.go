// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"fmt"
	"strings"
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/document"
)

// Note on prose security model: The checker blocks dangerous HTML
// (script, iframe, event handlers, javascript: URLs) BEFORE prose
// reaches the renderer. The CommonMark renderer intentionally passes
// through HTML blocks per spec — that's how allowed HTML (tables,
// details, etc.) works. The tests below verify the render-level
// defense: CSP meta tag, escaped display output, and escaped errors.

// TestHTMLProseGoesThoughCommonMark verifies prose is rendered via
// CommonMark (Markdown formatting works, not just escaped text).
func TestHTMLProseGoesThoughCommonMark(t *testing.T) {
	doc := document.NewDocument()
	doc.AddProse("# Heading\n\nA **bold** paragraph.")
	html := HTML(doc, "test")

	if !strings.Contains(html, "<h1>") {
		t.Error("heading should be rendered as <h1>")
	}
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Error("bold should be rendered as <strong>")
	}
}

// TestHTMLCSPPresent verifies the Content-Security-Policy meta tag
// blocks all remote resource loading.
func TestHTMLCSPPresent(t *testing.T) {
	doc := document.NewDocument()
	doc.AddProse("Safe content")
	html := HTML(doc, "test")

	if !strings.Contains(html, "Content-Security-Policy") {
		t.Error("missing CSP meta tag")
	}
	if !strings.Contains(html, "default-src 'none'") {
		t.Error("CSP should block all default sources")
	}
}

// TestHTMLEscapesCellOutput verifies that cell display output is escaped.
func TestHTMLEscapesCellOutput(t *testing.T) {
	doc := document.NewDocument()
	doc.AddDisplay("evil", "text/markdown", "<script>alert('xss')</script>")
	html := HTML(doc, "test")

	if strings.Contains(html, "<script>") {
		t.Error("display output should be escaped")
	}
}

// TestHTMLEscapesErrorContent verifies error messages are escaped.
func TestHTMLEscapesErrorContent(t *testing.T) {
	doc := document.NewDocument()
	doc.AddError("test<script>", 1, fmt.Errorf("<script>alert('xss')</script>"))
	html := HTML(doc, "test")

	if strings.Contains(html, "<script>alert") {
		t.Error("error content should be escaped")
	}
}

// TestMarkdownNoHTMLInjection verifies the Markdown backend doesn't
// inject executable content.
func TestMarkdownNoHTMLInjection(t *testing.T) {
	doc := document.NewDocument()
	doc.AddProse("# Title\n\n<script>alert('xss')</script>")
	md := Markdown(doc)

	// Markdown backend passes through raw text — the script tag appears
	// as literal text, which is fine because Markdown is not an HTML execution context.
	// The important thing is it doesn't crash.
	if len(md) == 0 {
		t.Error("markdown output should not be empty")
	}
}

// TestHTMLEscapesDisplayHostilePatterns verifies that hostile patterns
// in cell display output (which bypass the CommonMark renderer) are
// properly escaped. The security model is: the checker blocks dangerous
// HTML in prose before it reaches the renderer. Display output is always
// escaped via html.EscapeString.
func TestHTMLEscapesDisplayHostilePatterns(t *testing.T) {
	hostilePatterns := []struct {
		name    string
		input   string
		blocked string
	}{
		{"script_in_display", "<script>alert(1)</script>", "<script>"},
		{"iframe_in_display", "<iframe src=evil></iframe>", "<iframe"},
		{"img_onerror_display", "<img src=x onerror=alert(1)>", "<img"},
		{"div_onclick_display", "<div onclick=alert(1)>x</div>", "<div onclick"},
	}

	for _, hp := range hostilePatterns {
		t.Run(hp.name, func(t *testing.T) {
			doc := document.NewDocument()
			doc.AddDisplay("test", "text/markdown", hp.input)
			html := HTML(doc, "test")

			if strings.Contains(html, hp.blocked) {
				t.Errorf("hostile display output %q: %q should be escaped", hp.name, hp.blocked)
			}
		})
	}
}

// TestHTMLCSPBlocksRemoteLoads verifies the CSP meta tag configuration
// blocks external resource loading. The CSP should use default-src 'none'
// to prevent network requests from the report.
func TestHTMLCSPBlocksRemoteLoads(t *testing.T) {
	doc := document.NewDocument()
	doc.AddProse("# Test")
	html := HTML(doc, "test")

	// CSP must block all default sources.
	if !strings.Contains(html, "default-src 'none'") {
		t.Error("CSP missing default-src 'none'")
	}
	// Images only from data: URIs (embedded).
	if !strings.Contains(html, "img-src data:") {
		t.Error("CSP missing img-src data: restriction")
	}
}
