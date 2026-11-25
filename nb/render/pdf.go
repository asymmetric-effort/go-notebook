// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package render

import (
	"encoding/base64"
	"fmt"
	"regexp"
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
			} else if e.Mime == "image/svg+xml" {
				// SVG charts: embed as decoded image for Typst rendering.
				buf.WriteString(fmt.Sprintf("#image.decode(bytes((%s)), format: \"svg\")\n\n", svgToTypstBytes(e.Content)))
			} else if strings.HasPrefix(e.Mime, "image/") {
				// Raster images: embed as decoded image.
				buf.WriteString(fmt.Sprintf("#image.decode(bytes((%s)))\n\n", base64ToTypstBytes(e.Content)))
			}
		case document.KindError:
			buf.WriteString(fmt.Sprintf("#block(fill: rgb(\"#fee\"), stroke: rgb(\"#c00\"), inset: 1em)[\n"))
			buf.WriteString(fmt.Sprintf("  *Error in cell %q (line %d):* %s\n", e.CellName, e.Line, e.Content))
			buf.WriteString("]\n\n")
		case document.KindImage:
			if e.Mime == "image/svg+xml" {
				buf.WriteString(fmt.Sprintf("#image.decode(bytes((%s)), format: \"svg\")\n\n", svgToTypstBytes(e.Content)))
			} else {
				buf.WriteString(fmt.Sprintf("#image.decode(bytes((%s)))\n\n", base64ToTypstBytes(e.Content)))
			}
		}
	}

	return buf.String()
}

// svgToTypstBytes converts SVG string content to a Typst bytes literal.
// The content is the raw SVG string (from chart rendering).
func svgToTypstBytes(svgContent string) string {
	// Typst bytes literal: comma-separated byte values.
	var parts []string
	for _, b := range []byte(svgContent) {
		parts = append(parts, fmt.Sprintf("%d", b))
	}
	return strings.Join(parts, ", ")
}

// base64ToTypstBytes decodes base64 content to a Typst bytes literal.
func base64ToTypstBytes(b64Content string) string {
	data, err := base64.StdEncoding.DecodeString(b64Content)
	if err != nil {
		return "0" // Fallback for invalid base64.
	}
	var parts []string
	for _, b := range data {
		parts = append(parts, fmt.Sprintf("%d", b))
	}
	return strings.Join(parts, ", ")
}

// markdownToTypst converts Markdown prose to Typst markup. Handles headings
// (H1-H6), unordered/ordered lists, fenced code blocks, blockquotes,
// thematic breaks, emphasis, strong, code spans, and links.
func markdownToTypst(md string) string {
	lines := strings.Split(md, "\n")
	var result []string
	inCodeBlock := false

	inTable := false

	for _, line := range lines {
		// Fenced code blocks pass through verbatim.
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if inCodeBlock {
				result = append(result, "```")
				inCodeBlock = false
			} else {
				result = append(result, "```")
				inCodeBlock = true
			}
			continue
		}
		if inCodeBlock {
			result = append(result, line)
			continue
		}

		// Indented code blocks (4 spaces or 1 tab).
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			trimmed := line
			if strings.HasPrefix(trimmed, "    ") {
				trimmed = trimmed[4:]
			} else {
				trimmed = trimmed[1:]
			}
			result = append(result, fmt.Sprintf("#raw(\"%s\")", strings.ReplaceAll(trimmed, "\"", "\\\"")))
			continue
		}

		// Tables: pipe-delimited rows.
		if strings.Contains(line, "|") && strings.TrimSpace(line) != "" {
			cells := splitTableRow(line)
			if len(cells) > 0 {
				// Skip separator rows (---).
				if isTableSeparator(line) {
					continue
				}
				if !inTable {
					result = append(result, fmt.Sprintf("#table(columns: %d,", len(cells)))
					inTable = true
				}
				for _, cell := range cells {
					result = append(result, fmt.Sprintf("  [%s],", typstInline(strings.TrimSpace(cell))))
				}
				continue
			}
		}
		if inTable && !strings.Contains(line, "|") {
			result = append(result, ")")
			inTable = false
		}

		// Hard line breaks (two trailing spaces or backslash).
		if strings.HasSuffix(line, "  ") || strings.HasSuffix(line, "\\") {
			cleaned := strings.TrimRight(line, " \\")
			result = append(result, typstInline(cleaned)+" \\")
			continue
		}

		// ATX headings.
		if strings.HasPrefix(line, "###### ") {
			result = append(result, fmt.Sprintf("====== %s", typstInline(line[7:])))
		} else if strings.HasPrefix(line, "##### ") {
			result = append(result, fmt.Sprintf("===== %s", typstInline(line[6:])))
		} else if strings.HasPrefix(line, "#### ") {
			result = append(result, fmt.Sprintf("==== %s", typstInline(line[5:])))
		} else if strings.HasPrefix(line, "### ") {
			result = append(result, fmt.Sprintf("=== %s", typstInline(line[4:])))
		} else if strings.HasPrefix(line, "## ") {
			result = append(result, fmt.Sprintf("== %s", typstInline(line[3:])))
		} else if strings.HasPrefix(line, "# ") {
			result = append(result, fmt.Sprintf("= %s", typstInline(line[2:])))
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ") {
			// Unordered list items.
			result = append(result, fmt.Sprintf("- %s", typstInline(line[2:])))
		} else if len(line) > 2 && line[0] >= '0' && line[0] <= '9' && (strings.Contains(line[:3], ". ") || strings.Contains(line[:4], ". ")) {
			// Ordered list items.
			idx := strings.Index(line, ". ")
			if idx > 0 && idx < 10 {
				result = append(result, fmt.Sprintf("+ %s", typstInline(line[idx+2:])))
			} else {
				result = append(result, typstInline(line))
			}
		} else if strings.HasPrefix(line, "> ") {
			// Blockquote.
			result = append(result, fmt.Sprintf("#quote[%s]", typstInline(line[2:])))
		} else if isTypstThematicBreak(line) {
			result = append(result, "#line(length: 100%)")
		} else {
			result = append(result, typstInline(line))
		}
	}

	if inTable {
		result = append(result, ")")
	}

	return strings.Join(result, "\n")
}

// splitTableRow splits a Markdown table row into cells.
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "|") {
		line = line[1:]
	}
	if strings.HasSuffix(line, "|") {
		line = line[:len(line)-1]
	}
	if line == "" {
		return nil
	}
	return strings.Split(line, "|")
}

// isTableSeparator checks if a line is a table separator (e.g., |---|---|).
func isTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.ReplaceAll(trimmed, "|", "")
	trimmed = strings.ReplaceAll(trimmed, "-", "")
	trimmed = strings.ReplaceAll(trimmed, ":", "")
	trimmed = strings.TrimSpace(trimmed)
	return trimmed == ""
}

// typstInline converts inline Markdown elements to Typst: emphasis, strong,
// code spans, and links.
func typstInline(text string) string {
	// Code spans: `code` -> #raw("code")
	text = typstCodeSpanRe.ReplaceAllStringFunc(text, func(s string) string {
		code := strings.Trim(s, "`")
		return fmt.Sprintf("#raw(\"%s\")", strings.ReplaceAll(code, "\"", "\\\""))
	})
	// Images must be before links (![...] before [...]).
	text = typstImageRe.ReplaceAllString(text, "#image(\"${2}\")")
	// Links: [text](url) -> #link("url")[text]
	text = typstLinkRe.ReplaceAllString(text, "#link(\"${2}\")[${1}]")
	// Autolinks: <url> -> #link("url")
	text = typstAutolinkRe.ReplaceAllString(text, "#link(\"${1}\")")
	// Strong must be before emphasis (** before *).
	text = typstStrongStarRe.ReplaceAllString(text, "*${1}*")
	text = typstStrongUnderRe.ReplaceAllString(text, "*${1}*")
	return text
}

var typstCodeSpanRe = regexp.MustCompile("`[^`]+`")
var typstAutolinkRe = regexp.MustCompile(`<(https?://[^>]+)>`)
var typstStrongStarRe = regexp.MustCompile(`\*\*([^*]+)\*\*`)
var typstStrongUnderRe = regexp.MustCompile(`__([^_]+)__`)
var typstLinkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
var typstImageRe = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

func isTypstThematicBreak(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	ch := trimmed[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}
	for _, c := range trimmed {
		if c != rune(ch) && c != ' ' {
			return false
		}
	}
	return true
}
