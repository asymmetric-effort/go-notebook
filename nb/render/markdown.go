// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package render provides output backends for rendering the document model
// to Markdown, HTML and PDF.
package render

import (
	"fmt"
	"strings"

	"github.com/asymmetric-effort/go-notebook/nb/document"
)

// Markdown renders a document to Markdown format. Display output is inlined.
// Images are referenced by path (written to a directory next to the output).
func Markdown(doc *document.Document) string {
	var buf strings.Builder

	for _, e := range doc.Elements {
		switch e.Kind {
		case document.KindProse:
			buf.WriteString(e.Content)
			buf.WriteString("\n\n")
		case document.KindDisplay:
			if e.Mime == "text/markdown" {
				buf.WriteString(e.Content)
				buf.WriteString("\n\n")
			} else if strings.HasPrefix(e.Mime, "image/") {
				buf.WriteString(fmt.Sprintf("![%s](%s)\n\n", e.CellName, e.ImageRef))
			}
		case document.KindError:
			buf.WriteString(fmt.Sprintf("**Error in cell %q (line %d):** %s\n\n", e.CellName, e.Line, e.Content))
		case document.KindImage:
			buf.WriteString(fmt.Sprintf("![image](%s)\n\n", e.ImageRef))
		}
	}

	return strings.TrimRight(buf.String(), "\n") + "\n"
}
