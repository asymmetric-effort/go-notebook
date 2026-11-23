// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/document"
)

// FuzzRenderHTML fuzzes the HTML render pipeline with random prose content.
// It must never panic, regardless of input.
func FuzzRenderHTML(f *testing.F) {
	f.Add("# Hello World")
	f.Add("<script>alert('xss')</script>")
	f.Add("**bold** and *italic*")
	f.Add("")
	f.Add("<div onclick=\"evil()\">click</div>")
	f.Add("![image](../../../etc/passwd)")
	f.Add("`code` and ```block```")
	f.Add("<iframe src=evil></iframe>")
	f.Add("Normal paragraph with no special chars.")
	f.Add("<!-- HTML comment -->")

	f.Fuzz(func(t *testing.T, prose string) {
		doc := document.NewDocument()
		doc.AddProse(prose)
		result := HTML(doc, "fuzz-test")
		if len(result) == 0 {
			t.Error("HTML output should never be empty")
		}
	})
}

// FuzzRenderMarkdown fuzzes the Markdown render pipeline.
func FuzzRenderMarkdown(f *testing.F) {
	f.Add("# Hello")
	f.Add("<script>alert(1)</script>")
	f.Add("")
	f.Add("**bold**")

	f.Fuzz(func(t *testing.T, prose string) {
		doc := document.NewDocument()
		doc.AddProse(prose)
		_ = Markdown(doc)
		// Must not panic.
	})
}
