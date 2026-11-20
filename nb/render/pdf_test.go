// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"errors"
	"strings"
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/document"
)

func TestTypstProse(t *testing.T) {
	doc := document.NewDocument()
	doc.AddProse("# Title\n\nSome text.")
	result := Typst(doc)
	if !strings.Contains(result, "= Title") {
		t.Errorf("heading not converted:\n%s", result)
	}
}

func TestTypstDisplay(t *testing.T) {
	doc := document.NewDocument()
	doc.AddDisplay("test", "text/markdown", "- item 1")
	result := Typst(doc)
	if !strings.Contains(result, "- item 1") {
		t.Errorf("display not included:\n%s", result)
	}
}

func TestTypstError(t *testing.T) {
	doc := document.NewDocument()
	doc.AddError("load", 5, errors.New("fail"))
	result := Typst(doc)
	if !strings.Contains(result, "Error") {
		t.Errorf("error not included:\n%s", result)
	}
}

func TestTypstImage(t *testing.T) {
	doc := document.NewDocument()
	doc.AddImage("test.png", "image/png", "base64data")
	result := Typst(doc)
	if !strings.Contains(result, "#image") {
		t.Errorf("image not included:\n%s", result)
	}
}

func TestTypstDisplayImage(t *testing.T) {
	doc := document.NewDocument()
	doc.AddDisplay("chart", "image/png", "imgdata")
	result := Typst(doc)
	if !strings.Contains(result, "#image") {
		t.Errorf("display image not included:\n%s", result)
	}
}

func TestMarkdownToTypstHeadings(t *testing.T) {
	tests := []struct {
		md   string
		want string
	}{
		{"# Title", "= Title"},
		{"## Section", "== Section"},
		{"### Sub", "=== Sub"},
		{"- item", "- item"},
		{"plain text", "plain text"},
	}
	for _, tt := range tests {
		got := markdownToTypst(tt.md)
		if got != tt.want {
			t.Errorf("markdownToTypst(%q) = %q, want %q", tt.md, got, tt.want)
		}
	}
}

func TestTypstPreamble(t *testing.T) {
	doc := document.NewDocument()
	result := Typst(doc)
	if !strings.Contains(result, "#set page") {
		t.Error("missing page setup")
	}
	if !strings.Contains(result, "#set text") {
		t.Error("missing text setup")
	}
}
