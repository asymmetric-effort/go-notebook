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
		{"#### H4", "==== H4"},
		{"##### H5", "===== H5"},
		{"###### H6", "====== H6"},
		{"- item", "- item"},
		{"* item", "- item"},
		{"+ item", "- item"},
		{"plain text", "plain text"},
	}
	for _, tt := range tests {
		got := markdownToTypst(tt.md)
		if got != tt.want {
			t.Errorf("markdownToTypst(%q) = %q, want %q", tt.md, got, tt.want)
		}
	}
}

func TestMarkdownToTypstOrderedList(t *testing.T) {
	got := markdownToTypst("1. first\n2. second")
	if !strings.Contains(got, "+ first") {
		t.Errorf("ordered list: %q", got)
	}
}

func TestMarkdownToTypstBlockquote(t *testing.T) {
	got := markdownToTypst("> quoted text")
	if !strings.Contains(got, "#quote[quoted text]") {
		t.Errorf("blockquote: %q", got)
	}
}

func TestMarkdownToTypstThematicBreak(t *testing.T) {
	for _, input := range []string{"---", "***", "___"} {
		got := markdownToTypst(input)
		if !strings.Contains(got, "#line(length: 100%)") {
			t.Errorf("thematic break %q: %q", input, got)
		}
	}
}

func TestMarkdownToTypstCodeBlock(t *testing.T) {
	got := markdownToTypst("```\ncode\n```")
	if !strings.Contains(got, "```") || !strings.Contains(got, "code") {
		t.Errorf("code block: %q", got)
	}
}

func TestMarkdownToTypstInlineCode(t *testing.T) {
	got := markdownToTypst("Use `fmt.Println` here")
	if !strings.Contains(got, "#raw(\"fmt.Println\")") {
		t.Errorf("inline code: %q", got)
	}
}

func TestMarkdownToTypstStrong(t *testing.T) {
	got := markdownToTypst("**bold**")
	if !strings.Contains(got, "*bold*") {
		t.Errorf("strong: %q", got)
	}
}

func TestMarkdownToTypstLink(t *testing.T) {
	got := markdownToTypst("[Go](https://go.dev)")
	if !strings.Contains(got, "#link(\"https://go.dev\")[Go]") {
		t.Errorf("link: %q", got)
	}
}

func TestMarkdownToTypstImage(t *testing.T) {
	got := markdownToTypst("![alt](image.png)")
	if !strings.Contains(got, "#image(\"image.png\")") {
		t.Errorf("image: %q", got)
	}
}

func TestIsTypstThematicBreak(t *testing.T) {
	if !isTypstThematicBreak("---") {
		t.Error("--- should be thematic break")
	}
	if isTypstThematicBreak("--") {
		t.Error("-- should not be thematic break")
	}
	if isTypstThematicBreak("-*-") {
		t.Error("mixed chars should not be thematic break")
	}
}

func TestTypstSVGChart(t *testing.T) {
	doc := document.NewDocument()
	doc.AddDisplay("chart", "image/svg+xml", "<svg><rect/></svg>")
	result := Typst(doc)
	if !strings.Contains(result, "image.decode") {
		t.Error("SVG chart should use image.decode")
	}
	if !strings.Contains(result, "format: \"svg\"") {
		t.Error("SVG chart should specify svg format")
	}
}

func TestTypstRasterImage(t *testing.T) {
	doc := document.NewDocument()
	doc.AddImage("test.png", "image/png", "iVBORw0KGgo=")
	result := Typst(doc)
	if !strings.Contains(result, "image.decode") {
		t.Error("raster image should use image.decode")
	}
}

func TestSvgToTypstBytes(t *testing.T) {
	result := svgToTypstBytes("AB")
	if result != "65, 66" {
		t.Errorf("svgToTypstBytes(\"AB\") = %q, want \"65, 66\"", result)
	}
}

func TestBase64ToTypstBytes(t *testing.T) {
	// "hi" in base64 is "aGk="
	result := base64ToTypstBytes("aGk=")
	if result != "104, 105" {
		t.Errorf("base64ToTypstBytes = %q, want \"104, 105\"", result)
	}
}

func TestBase64ToTypstBytesInvalid(t *testing.T) {
	result := base64ToTypstBytes("!!!invalid!!!")
	if result != "0" {
		t.Errorf("invalid base64 = %q, want \"0\"", result)
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
