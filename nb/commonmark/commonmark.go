// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package commonmark implements a zero-dependency CommonMark-to-HTML renderer.
// It handles the most common Markdown constructs: headings, paragraphs,
// emphasis, strong, code spans, links, images, lists, blockquotes, fenced
// code blocks, thematic breaks, and hard line breaks.
package commonmark

import (
	"html"
	"regexp"
	"strings"
)

// ToHTML converts CommonMark Markdown to HTML.
func ToHTML(src string) string {
	lines := strings.Split(src, "\n")
	blocks := parseBlocks(lines)
	var buf strings.Builder
	for _, b := range blocks {
		renderBlock(&buf, b)
	}
	return buf.String()
}

// blockKind identifies the type of a block-level element.
type blockKind int

const (
	blockParagraph blockKind = iota
	blockHeading
	blockFencedCode
	blockThematicBreak
	blockBlockquote
	blockUnorderedList
	blockOrderedList
	blockHTML
)

type block struct {
	kind    blockKind
	level   int      // heading level (1-6) or list start number
	content string   // text content
	info    string   // fenced code info string
	items   []string // list items
}

// parseBlocks splits input lines into block-level elements.
func parseBlocks(lines []string) []block {
	var blocks []block
	i := 0
	for i < len(lines) {
		line := lines[i]

		// Blank line — skip.
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// ATX heading.
		if level, text, ok := parseATXHeading(line); ok {
			blocks = append(blocks, block{kind: blockHeading, level: level, content: text})
			i++
			continue
		}

		// Thematic break.
		if isThematicBreak(line) {
			blocks = append(blocks, block{kind: blockThematicBreak})
			i++
			continue
		}

		// Fenced code block.
		if fence, info, ok := parseFenceOpener(line); ok {
			var codeLines []string
			i++
			for i < len(lines) {
				if isFenceCloser(lines[i], fence) {
					i++
					break
				}
				codeLines = append(codeLines, lines[i])
				i++
			}
			blocks = append(blocks, block{kind: blockFencedCode, content: strings.Join(codeLines, "\n"), info: info})
			continue
		}

		// Blockquote.
		if strings.HasPrefix(strings.TrimLeft(line, " "), "> ") || strings.TrimSpace(line) == ">" {
			var quoteLines []string
			for i < len(lines) {
				trimmed := strings.TrimLeft(lines[i], " ")
				if strings.HasPrefix(trimmed, "> ") {
					quoteLines = append(quoteLines, trimmed[2:])
				} else if strings.TrimSpace(trimmed) == ">" {
					quoteLines = append(quoteLines, "")
				} else {
					break
				}
				i++
			}
			inner := ToHTML(strings.Join(quoteLines, "\n"))
			blocks = append(blocks, block{kind: blockBlockquote, content: inner})
			continue
		}

		// Unordered list.
		if isUnorderedListItem(line) {
			var items []string
			for i < len(lines) {
				if isUnorderedListItem(lines[i]) {
					items = append(items, stripListMarker(lines[i]))
				} else if strings.TrimSpace(lines[i]) == "" {
					break
				} else {
					// Continuation line.
					if len(items) > 0 {
						items[len(items)-1] += "\n" + strings.TrimSpace(lines[i])
					}
				}
				i++
			}
			blocks = append(blocks, block{kind: blockUnorderedList, items: items})
			continue
		}

		// Ordered list.
		if isOrderedListItem(line) {
			var items []string
			for i < len(lines) {
				if isOrderedListItem(lines[i]) {
					items = append(items, stripOrderedMarker(lines[i]))
				} else if strings.TrimSpace(lines[i]) == "" {
					break
				} else {
					if len(items) > 0 {
						items[len(items)-1] += "\n" + strings.TrimSpace(lines[i])
					}
				}
				i++
			}
			blocks = append(blocks, block{kind: blockOrderedList, items: items})
			continue
		}

		// Raw HTML block (starts with < followed by a tag name, not an autolink).
		trimmed := strings.TrimSpace(line)
		if isHTMLBlockStart(trimmed) {
			var htmlLines []string
			for i < len(lines) {
				if strings.TrimSpace(lines[i]) == "" {
					break
				}
				htmlLines = append(htmlLines, lines[i])
				i++
			}
			blocks = append(blocks, block{kind: blockHTML, content: strings.Join(htmlLines, "\n")})
			continue
		}

		// Paragraph — collect lines until blank line or block start.
		var paraLines []string
		for i < len(lines) {
			l := lines[i]
			if strings.TrimSpace(l) == "" {
				break
			}
			if _, _, ok := parseATXHeading(l); ok {
				break
			}
			if isThematicBreak(l) {
				break
			}
			if _, _, ok := parseFenceOpener(l); ok {
				break
			}
			if strings.HasPrefix(strings.TrimLeft(l, " "), "> ") {
				break
			}
			if isUnorderedListItem(l) && len(paraLines) > 0 {
				break
			}
			if isOrderedListItem(l) && len(paraLines) > 0 {
				break
			}
			paraLines = append(paraLines, l)
			i++
		}
		if len(paraLines) > 0 {
			blocks = append(blocks, block{kind: blockParagraph, content: strings.Join(paraLines, "\n")})
		}
	}
	return blocks
}

func renderBlock(buf *strings.Builder, b block) {
	switch b.kind {
	case blockHeading:
		buf.WriteString("<h")
		buf.WriteByte('0' + byte(b.level))
		buf.WriteString(">")
		buf.WriteString(renderInline(b.content))
		buf.WriteString("</h")
		buf.WriteByte('0' + byte(b.level))
		buf.WriteString(">\n")
	case blockParagraph:
		buf.WriteString("<p>")
		buf.WriteString(renderInline(b.content))
		buf.WriteString("</p>\n")
	case blockFencedCode:
		if b.info != "" {
			lang := strings.Fields(b.info)[0]
			buf.WriteString("<pre><code class=\"language-")
			buf.WriteString(html.EscapeString(lang))
			buf.WriteString("\">")
		} else {
			buf.WriteString("<pre><code>")
		}
		buf.WriteString(html.EscapeString(b.content))
		buf.WriteString("</code></pre>\n")
	case blockThematicBreak:
		buf.WriteString("<hr />\n")
	case blockBlockquote:
		buf.WriteString("<blockquote>\n")
		buf.WriteString(b.content)
		buf.WriteString("</blockquote>\n")
	case blockUnorderedList:
		buf.WriteString("<ul>\n")
		for _, item := range b.items {
			buf.WriteString("<li>")
			buf.WriteString(renderInline(item))
			buf.WriteString("</li>\n")
		}
		buf.WriteString("</ul>\n")
	case blockOrderedList:
		buf.WriteString("<ol>\n")
		for _, item := range b.items {
			buf.WriteString("<li>")
			buf.WriteString(renderInline(item))
			buf.WriteString("</li>\n")
		}
		buf.WriteString("</ol>\n")
	case blockHTML:
		buf.WriteString(b.content)
		buf.WriteString("\n")
	}
}

// --- Block-level helpers ---

func parseATXHeading(line string) (int, string, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(trimmed) == 0 || trimmed[0] != '#' {
		return 0, "", false
	}
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level > 6 {
		return 0, "", false
	}
	if level == len(trimmed) {
		return level, "", true // empty heading
	}
	if trimmed[level] != ' ' {
		return 0, "", false
	}
	text := strings.TrimSpace(trimmed[level+1:])
	// Remove trailing # characters.
	text = strings.TrimRight(text, " #")
	return level, text, true
}

func isThematicBreak(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	ch := trimmed[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}
	count := 0
	for _, c := range trimmed {
		if c == rune(ch) {
			count++
		} else if c != ' ' {
			return false
		}
	}
	return count >= 3
}

func parseFenceOpener(line string) (string, string, bool) {
	trimmed := strings.TrimLeft(line, " ")
	if len(trimmed) < 3 {
		return "", "", false
	}
	ch := trimmed[0]
	if ch != '`' && ch != '~' {
		return "", "", false
	}
	count := 0
	for count < len(trimmed) && trimmed[count] == ch {
		count++
	}
	if count < 3 {
		return "", "", false
	}
	fence := string(ch) + strings.Repeat(string(ch), count-1)
	info := strings.TrimSpace(trimmed[count:])
	return fence, info, true
}

func isFenceCloser(line string, fence string) bool {
	trimmed := strings.TrimLeft(line, " ")
	trimmed = strings.TrimSpace(trimmed)
	if len(trimmed) < len(fence) {
		return false
	}
	ch := fence[0]
	count := 0
	for count < len(trimmed) && trimmed[count] == ch {
		count++
	}
	return count >= len(fence) && strings.TrimSpace(trimmed[count:]) == ""
}

// isHTMLBlockStart checks if a line starts an HTML block (not an autolink or inline element).
func isHTMLBlockStart(line string) bool {
	if !strings.HasPrefix(line, "<") {
		return false
	}
	// Autolinks like <https://...> are not HTML blocks.
	if len(line) > 1 && (line[1] >= 'a' && line[1] <= 'z' || line[1] >= 'A' && line[1] <= 'Z') {
		// Check for known block-level HTML tags.
		lower := strings.ToLower(line)
		for _, tag := range []string{"<div", "<p ", "<p>", "<table", "<pre", "<blockquote",
			"<h1", "<h2", "<h3", "<h4", "<h5", "<h6", "<ul", "<ol", "<li",
			"<hr", "<section", "<article", "<nav", "<header", "<footer",
			"<details", "<summary", "<figure", "<figcaption",
			"<!doctype", "<!--", "</div", "</p>", "</table", "</pre",
			"<style", "<link"} {
			if strings.HasPrefix(lower, tag) {
				return true
			}
		}
	}
	if strings.HasPrefix(line, "<!--") {
		return true
	}
	return false
}

var unorderedListRe = regexp.MustCompile(`^[ ]{0,3}[-*+] `)
var orderedListRe = regexp.MustCompile(`^[ ]{0,3}\d{1,9}[.)] `)

func isUnorderedListItem(line string) bool {
	return unorderedListRe.MatchString(line)
}

func isOrderedListItem(line string) bool {
	return orderedListRe.MatchString(line)
}

func stripListMarker(line string) string {
	loc := unorderedListRe.FindStringIndex(line)
	if loc == nil {
		return line
	}
	return strings.TrimSpace(line[loc[1]:])
}

func stripOrderedMarker(line string) string {
	loc := orderedListRe.FindStringIndex(line)
	if loc == nil {
		return line
	}
	return strings.TrimSpace(line[loc[1]:])
}

// --- Inline rendering ---

// renderInline converts inline Markdown to HTML.
func renderInline(text string) string {
	// Handle hard line breaks (two trailing spaces + newline, or backslash + newline).
	text = strings.ReplaceAll(text, "  \n", "<br />\n")
	text = strings.ReplaceAll(text, "\\\n", "<br />\n")

	var buf strings.Builder
	i := 0
	runes := []rune(text)
	for i < len(runes) {
		ch := runes[i]

		// Backslash escape.
		if ch == '\\' && i+1 < len(runes) {
			next := runes[i+1]
			if isPunctuation(next) {
				buf.WriteString(html.EscapeString(string(next)))
				i += 2
				continue
			}
		}

		// Code span.
		if ch == '`' {
			end := findCodeSpanEnd(runes, i)
			if end > i {
				code := string(runes[i+1 : end])
				// Strip one space from each end if both are present.
				if len(code) > 1 && code[0] == ' ' && code[len(code)-1] == ' ' {
					code = code[1 : len(code)-1]
				}
				buf.WriteString("<code>")
				buf.WriteString(html.EscapeString(code))
				buf.WriteString("</code>")
				i = end + 1
				continue
			}
		}

		// Image: ![alt](src)
		if ch == '!' && i+1 < len(runes) && runes[i+1] == '[' {
			if alt, src, end, ok := parseLinkOrImage(runes, i+1); ok {
				buf.WriteString("<img src=\"")
				buf.WriteString(html.EscapeString(src))
				buf.WriteString("\" alt=\"")
				buf.WriteString(html.EscapeString(alt))
				buf.WriteString("\" />")
				i = end
				continue
			}
		}

		// Link: [text](url)
		if ch == '[' {
			if linkText, href, end, ok := parseLinkOrImage(runes, i); ok {
				buf.WriteString("<a href=\"")
				buf.WriteString(html.EscapeString(href))
				buf.WriteString("\">")
				buf.WriteString(renderInline(linkText))
				buf.WriteString("</a>")
				i = end
				continue
			}
		}

		// Strong: ** or __
		if (ch == '*' || ch == '_') && i+1 < len(runes) && runes[i+1] == ch {
			end := findDelimiterEnd(runes, i+2, ch, 2)
			if end > 0 {
				inner := string(runes[i+2 : end])
				buf.WriteString("<strong>")
				buf.WriteString(renderInline(inner))
				buf.WriteString("</strong>")
				i = end + 2
				continue
			}
		}

		// Emphasis: * or _
		if ch == '*' || ch == '_' {
			end := findDelimiterEnd(runes, i+1, ch, 1)
			if end > 0 {
				inner := string(runes[i+1 : end])
				buf.WriteString("<em>")
				buf.WriteString(renderInline(inner))
				buf.WriteString("</em>")
				i = end + 1
				continue
			}
		}

		// Autolinks: <url>
		if ch == '<' {
			if end := findAutolink(runes, i); end > 0 {
				url := string(runes[i+1 : end])
				buf.WriteString("<a href=\"")
				buf.WriteString(html.EscapeString(url))
				buf.WriteString("\">")
				buf.WriteString(html.EscapeString(url))
				buf.WriteString("</a>")
				i = end + 1
				continue
			}
		}

		// HTML entities and special characters.
		if ch == '&' {
			buf.WriteString("&amp;")
		} else if ch == '<' {
			buf.WriteString("&lt;")
		} else if ch == '>' {
			buf.WriteString("&gt;")
		} else if ch == '"' {
			buf.WriteString("&quot;")
		} else {
			buf.WriteRune(ch)
		}
		i++
	}
	return buf.String()
}

func isPunctuation(ch rune) bool {
	return strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", ch)
}

func findCodeSpanEnd(runes []rune, start int) int {
	// Count opening backticks.
	count := 0
	i := start
	for i < len(runes) && runes[i] == '`' {
		count++
		i++
	}
	// Find matching closing backticks.
	for i < len(runes) {
		j := i
		closeCount := 0
		for j < len(runes) && runes[j] == '`' {
			closeCount++
			j++
		}
		if closeCount == count {
			return j - count
		}
		if closeCount > 0 {
			i = j
		} else {
			i++
		}
	}
	return -1
}

func parseLinkOrImage(runes []rune, start int) (string, string, int, bool) {
	// Find the matching ].
	if start >= len(runes) || runes[start] != '[' {
		return "", "", 0, false
	}
	depth := 0
	i := start
	for i < len(runes) {
		if runes[i] == '[' {
			depth++
		} else if runes[i] == ']' {
			depth--
			if depth == 0 {
				break
			}
		}
		i++
	}
	if i >= len(runes) || depth != 0 {
		return "", "", 0, false
	}
	text := string(runes[start+1 : i])
	i++ // skip ]
	if i >= len(runes) || runes[i] != '(' {
		return "", "", 0, false
	}
	i++ // skip (
	// Find closing ).
	parenStart := i
	for i < len(runes) && runes[i] != ')' {
		i++
	}
	if i >= len(runes) {
		return "", "", 0, false
	}
	href := strings.TrimSpace(string(runes[parenStart:i]))
	i++ // skip )
	return text, href, i, true
}

func findDelimiterEnd(runes []rune, start int, ch rune, count int) int {
	for i := start; i <= len(runes)-count; i++ {
		match := true
		for j := 0; j < count; j++ {
			if runes[i+j] != ch {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func findAutolink(runes []rune, start int) int {
	// Look for <scheme:...> pattern.
	i := start + 1
	hasScheme := false
	for i < len(runes) {
		if runes[i] == ':' {
			hasScheme = true
		} else if runes[i] == '>' {
			if hasScheme {
				return i
			}
			return -1
		} else if runes[i] == ' ' || runes[i] == '\n' {
			return -1
		}
		i++
	}
	return -1
}
