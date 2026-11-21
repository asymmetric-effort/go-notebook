// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"errors"
	"strings"
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/document"
)

func TestMarkdownProse(t *testing.T) {
	doc := document.NewDocument()
	doc.AddProse("# Title")
	doc.AddProse("Some text.")
	result := Markdown(doc)
	if !strings.Contains(result, "# Title") {
		t.Errorf("missing title:\n%s", result)
	}
}

func TestMarkdownDisplay(t *testing.T) {
	doc := document.NewDocument()
	doc.AddDisplay("test", "text/markdown", "- item 1")
	result := Markdown(doc)
	if !strings.Contains(result, "- item 1") {
		t.Errorf("missing display:\n%s", result)
	}
}

func TestMarkdownError(t *testing.T) {
	doc := document.NewDocument()
	doc.AddError("load", 42, errors.New("file not found"))
	result := Markdown(doc)
	if !strings.Contains(result, "Error") || !strings.Contains(result, "load") {
		t.Errorf("missing error:\n%s", result)
	}
}

func TestMarkdownImage(t *testing.T) {
	doc := document.NewDocument()
	doc.AddImage("logo.png", "image/png", "data")
	result := Markdown(doc)
	if !strings.Contains(result, "![image](logo.png)") {
		t.Errorf("missing image:\n%s", result)
	}
}

func TestMarkdownDisplayImage(t *testing.T) {
	doc := document.NewDocument()
	doc.AddDisplay("chart", "image/png", "base64data")
	result := Markdown(doc)
	if !strings.Contains(result, "![chart]") {
		t.Errorf("missing display image:\n%s", result)
	}
}

func TestHTMLBasic(t *testing.T) {
	doc := document.NewDocument()
	doc.AddProse("Hello world")
	result := HTML(doc, "Test Report")
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE")
	}
	if !strings.Contains(result, "Test Report") {
		t.Error("missing title")
	}
	if !strings.Contains(result, "Content-Security-Policy") {
		t.Error("missing CSP")
	}
	if !strings.Contains(result, "Hello world") {
		t.Error("missing content")
	}
}

func TestHTMLCSP(t *testing.T) {
	doc := document.NewDocument()
	result := HTML(doc, "Test")
	if !strings.Contains(result, "default-src 'none'") {
		t.Error("CSP should block all remote loads")
	}
}

func TestHTMLDisplay(t *testing.T) {
	doc := document.NewDocument()
	doc.AddDisplay("test", "text/markdown", "output")
	result := HTML(doc, "Test")
	if !strings.Contains(result, "output") {
		t.Error("missing display output")
	}
}

func TestHTMLError(t *testing.T) {
	doc := document.NewDocument()
	doc.AddError("load", 5, errors.New("fail"))
	result := HTML(doc, "Test")
	if !strings.Contains(result, "class=\"error\"") {
		t.Error("missing error styling")
	}
}

func TestHTMLImage(t *testing.T) {
	doc := document.NewDocument()
	doc.AddImage("test.png", "image/png", "base64data")
	result := HTML(doc, "Test")
	if !strings.Contains(result, "data:image/png;base64,base64data") {
		t.Error("missing data URI image")
	}
}

func TestHTMLDisplayImage(t *testing.T) {
	doc := document.NewDocument()
	doc.AddDisplay("chart", "image/png", "imgdata")
	result := HTML(doc, "Test")
	if !strings.Contains(result, "data:image/png;base64,imgdata") {
		t.Error("missing display image data URI")
	}
}

func TestHTMLEscaping(t *testing.T) {
	doc := document.NewDocument()
	doc.AddProse("<script>alert('xss')</script>")
	result := HTML(doc, "Test")
	if strings.Contains(result, "<script>") {
		t.Error("script tag should be escaped")
	}
}

func TestHTMLStandalone(t *testing.T) {
	// Verify the HTML output is fully self-contained with no external references.
	doc := document.NewDocument()
	doc.AddProse("# Report Title\n\nSome **bold** text.")
	doc.AddDisplay("chart", "image/svg+xml", "<svg></svg>")
	doc.AddDisplay("data", "text/markdown", "result: 42")
	doc.AddImage("logo.png", "image/png", "iVBORw0KGgo=")
	doc.AddError("load", 10, errors.New("file not found"))

	result := HTML(doc, "Full Report")

	// Must be valid HTML with all required sections.
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE")
	}
	if !strings.Contains(result, "<meta charset=\"utf-8\">") {
		t.Error("missing charset")
	}
	if !strings.Contains(result, "Content-Security-Policy") {
		t.Error("missing CSP meta tag")
	}
	if !strings.Contains(result, "default-src 'none'") {
		t.Error("CSP must block all remote loads by default")
	}
	if !strings.Contains(result, "style-src 'unsafe-inline'") {
		t.Error("CSP must allow inline styles")
	}
	if !strings.Contains(result, "img-src data:") {
		t.Error("CSP must allow data: images")
	}
	// Styles must be inline, not linked.
	if !strings.Contains(result, "<style>") {
		t.Error("styles must be inline")
	}
	if strings.Contains(result, "<link rel=\"stylesheet\"") {
		t.Error("no external stylesheets allowed")
	}
	// Images must be data URIs.
	if !strings.Contains(result, "data:image/png;base64,") {
		t.Error("images must be data URIs")
	}
	// No external references.
	if strings.Contains(result, "http://") || strings.Contains(result, "https://") {
		t.Error("no external URLs allowed in output")
	}
	// Prose should be rendered through CommonMark.
	if !strings.Contains(result, "<h1>Report Title</h1>") {
		t.Error("prose should render through CommonMark")
	}
	if !strings.Contains(result, "<strong>bold</strong>") {
		t.Error("emphasis should be rendered")
	}
}

func TestHTMLProseCommonMark(t *testing.T) {
	doc := document.NewDocument()
	doc.AddProse("- item 1\n- item 2")
	result := HTML(doc, "Test")
	if !strings.Contains(result, "<ul>") {
		t.Error("list should be rendered as HTML")
	}
	if !strings.Contains(result, "<li>item 1</li>") {
		t.Error("list items should be rendered")
	}
}

func TestMarkdownEmpty(t *testing.T) {
	doc := document.NewDocument()
	result := Markdown(doc)
	if result != "\n" {
		t.Errorf("empty doc = %q, want newline", result)
	}
}
