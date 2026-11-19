// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"fmt"
	"go/format"
	"strings"
)

// FormatCell formats a cell body using gofmt. Since gofmt cannot format a
// bare body, the formatter wraps the body in a function, formats it, and
// strips the wrapper.
func FormatCell(body string) (string, error) {
	wrapped := wrapBody(body)
	formatted, err := format.Source([]byte(wrapped))
	if err != nil {
		return "", fmt.Errorf("format cell: %w", err)
	}
	return unwrapBody(string(formatted))
}

// FormatDefs formats a go:defs block. Since defs blocks contain top-level
// declarations, they can be formatted directly as a Go source file.
func FormatDefs(body string) (string, error) {
	// Wrap in a package declaration so gofmt can parse it.
	src := "package _\n\n" + body + "\n"
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return "", fmt.Errorf("format defs: %w", err)
	}
	// Strip the "package _\n\n" prefix.
	result := strings.TrimPrefix(string(formatted), "package _\n\n")
	result = strings.TrimRight(result, "\n")
	return result, nil
}

// wrapBody wraps a cell body in a function so gofmt can format it.
func wrapBody(body string) string {
	return "package _\n\nfunc _() {\n" + indentBody(body) + "\n}\n"
}

// unwrapBody extracts the formatted body from the wrapper function.
func unwrapBody(formatted string) (string, error) {
	// Find the opening brace.
	start := strings.Index(formatted, "func _() {\n")
	if start < 0 {
		return "", fmt.Errorf("cannot find wrapper function in formatted output")
	}
	start += len("func _() {\n")

	// Find the closing brace (last line).
	end := strings.LastIndex(formatted, "\n}")
	if end < 0 || end < start {
		return "", fmt.Errorf("cannot find closing brace in formatted output")
	}

	body := formatted[start:end]
	// Remove one level of indentation.
	lines := strings.Split(body, "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, "\t") {
			lines[i] = l[1:]
		}
	}
	return strings.Join(lines, "\n"), nil
}

// FormatNotebook formats all go:notebook and go:defs blocks in a parsed
// notebook and returns the reformatted Markdown source.
func FormatNotebook(source string) (string, error) {
	lines := strings.Split(source, "\n")
	result := Parse(source)

	// Process blocks in reverse order so positions stay valid.
	for i := len(result.Blocks) - 1; i >= 0; i-- {
		b := result.Blocks[i]
		var formatted string
		var err error

		switch b.Kind {
		case BlockCell:
			formatted, err = FormatCell(b.Body)
		case BlockDefs:
			formatted, err = FormatDefs(b.Body)
		default:
			continue
		}

		if err != nil {
			return "", fmt.Errorf("line %d: %w", b.StartPos.Line, err)
		}

		if formatted == b.Body {
			continue
		}

		// Replace the body lines in the source.
		bodyStart := b.BodyPos.Line - 1 // 0-based
		bodyEnd := b.EndPos.Line - 1    // 0-based, exclusive (closing fence line)

		formattedLines := strings.Split(formatted, "\n")
		newLines := make([]string, 0, len(lines))
		newLines = append(newLines, lines[:bodyStart]...)
		newLines = append(newLines, formattedLines...)
		newLines = append(newLines, lines[bodyEnd:]...)
		lines = newLines
	}

	return strings.Join(lines, "\n"), nil
}
