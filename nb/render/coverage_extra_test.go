// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"strings"
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/document"
)

func TestTypstCodeBlock(t *testing.T) {
	// Tests the ``` branch in markdownToTypst.
	doc := document.NewDocument()
	doc.AddProse("```\ncode here\n```")
	result := Typst(doc)
	if !strings.Contains(result, "```") {
		t.Error("code block backticks should pass through")
	}
}

func TestMarkdownToTypstPlainText(t *testing.T) {
	result := markdownToTypst("just some plain text")
	if result != "just some plain text" {
		t.Errorf("plain text should pass through: %q", result)
	}
}
