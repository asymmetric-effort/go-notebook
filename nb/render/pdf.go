// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"fmt"
	"strings"

	"github.com/asymmetric-effort/go-notebook/nb/document"
)

// Typst converts a document model to Typst markup for PDF compilation.
func Typst(doc *document.Document) string {
	var buf strings.Builder

	// Typst preamble.
	buf.WriteString("#set page(margin: 2cm)\n")
	buf.WriteString("#set text(font: \"New Computer Modern\", size: 11pt)\n\n")

	for _, e := range doc.Elements {
		switch e.Kind {
		case document.KindProse:
			buf.WriteString(markdownToTypst(e.Content))
			buf.WriteString("\n\n")
		case document.KindDisplay:
			if e.Mime == "text/markdown" {
				buf.WriteString(markdownToTypst(e.Content))
				buf.WriteString("\n\n")
			} else if strings.HasPrefix(e.Mime, "image/") {
				// Images would be embedded as Typst image() calls.
				buf.WriteString(fmt.Sprintf("#image(\"data:%s;base64,%s\")\n\n", e.Mime, e.Content))
			}
		case document.KindError:
			buf.WriteString(fmt.Sprintf("#block(fill: rgb(\"#fee\"), stroke: rgb(\"#c00\"), inset: 1em)[\n"))
			buf.WriteString(fmt.Sprintf("  *Error in cell %q (line %d):* %s\n", e.CellName, e.Line, e.Content))
			buf.WriteString("]\n\n")
		case document.KindImage:
			buf.WriteString(fmt.Sprintf("#image(\"data:%s;base64,%s\")\n\n", e.Mime, e.Content))
		}
	}

	return buf.String()
}

// markdownToTypst does a basic conversion from Markdown to Typst markup.
// A full implementation would handle all CommonMark features.
func markdownToTypst(md string) string {
	lines := strings.Split(md, "\n")
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			result = append(result, fmt.Sprintf("= %s", line[2:]))
		} else if strings.HasPrefix(line, "## ") {
			result = append(result, fmt.Sprintf("== %s", line[3:]))
		} else if strings.HasPrefix(line, "### ") {
			result = append(result, fmt.Sprintf("=== %s", line[4:]))
		} else if strings.HasPrefix(line, "- ") {
			result = append(result, fmt.Sprintf("- %s", line[2:]))
		} else if strings.HasPrefix(line, "```") {
			result = append(result, "```")
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
