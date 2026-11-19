// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package core provides the parser, checker, generator and formatter for
// go-notebook. It is the shared library used by the CLI, server and LSP.
package core

import (
	"strings"
)

// BlockKind identifies the type of a parsed block.
type BlockKind int

const (
	// BlockProse is any non-cell Markdown content.
	BlockProse BlockKind = iota
	// BlockDefs is a go:defs fenced block.
	BlockDefs
	// BlockCell is a go:notebook fenced block.
	BlockCell
	// BlockOutput is an output fenced block (generated, never executed).
	BlockOutput
)

// String returns a human-readable name for a BlockKind.
func (k BlockKind) String() string {
	switch k {
	case BlockProse:
		return "prose"
	case BlockDefs:
		return "defs"
	case BlockCell:
		return "cell"
	case BlockOutput:
		return "output"
	default:
		return "unknown"
	}
}

// Position identifies a location in the source Markdown file.
type Position struct {
	Line int // 1-based line number.
	Col  int // 1-based column number.
}

// Block is a parsed segment of a notebook Markdown file.
type Block struct {
	Kind     BlockKind
	Info     string   // Full info string for fenced blocks (e.g., "go:notebook name=load").
	Body     string   // The content between the fences, excluding the fence lines.
	StartPos Position // Position of the opening fence (or first line of prose).
	EndPos   Position // Position of the closing fence (or last line of prose).
	BodyPos  Position // Position of the first line of the body content.
}

// ParseResult holds the output of parsing a notebook Markdown file.
type ParseResult struct {
	Blocks []Block
}

// Parse extracts blocks from a notebook Markdown file. It identifies go:defs,
// go:notebook and output fenced code blocks, treating everything else as prose.
// Fenced blocks use the CommonMark rules: at least 3 backticks or tildes,
// optionally indented up to 3 spaces. A closing fence must use the same
// character, be at least as long, and have no info string.
func Parse(source string) ParseResult {
	lines := splitLines(source)
	var blocks []Block
	var proseLines []string
	proseStart := Position{Line: 1, Col: 1}

	i := 0
	for i < len(lines) {
		fenceChar, fenceLen, indent, info := parseFenceOpening(lines[i])
		if fenceLen > 0 {
			// Flush accumulated prose.
			if len(proseLines) > 0 {
				blocks = append(blocks, Block{
					Kind:     BlockProse,
					Body:     strings.Join(proseLines, "\n"),
					StartPos: proseStart,
					EndPos:   Position{Line: i, Col: 1},
				})
				proseLines = nil
			}

			kind := classifyInfoString(info)
			openLine := i + 1 // 1-based
			bodyStart := i + 1
			i++

			// Find closing fence.
			var bodyLines []string
			closed := false
			for i < len(lines) {
				if isClosingFence(lines[i], fenceChar, fenceLen, indent) {
					closed = true
					break
				}
				bodyLine := removeIndent(lines[i], indent)
				bodyLines = append(bodyLines, bodyLine)
				i++
			}

			closeLine := i + 1 // 1-based
			if !closed {
				closeLine = len(lines)
			}

			blocks = append(blocks, Block{
				Kind:     kind,
				Info:     info,
				Body:     strings.Join(bodyLines, "\n"),
				StartPos: Position{Line: openLine, Col: indent + 1},
				EndPos:   Position{Line: closeLine, Col: indent + 1},
				BodyPos:  Position{Line: bodyStart + 1, Col: indent + 1},
			})

			if closed {
				i++ // Skip closing fence.
			}
			proseStart = Position{Line: i + 1, Col: 1}
		} else {
			if len(proseLines) == 0 {
				proseStart = Position{Line: i + 1, Col: 1}
			}
			proseLines = append(proseLines, lines[i])
			i++
		}
	}

	// Flush trailing prose.
	if len(proseLines) > 0 {
		blocks = append(blocks, Block{
			Kind:     BlockProse,
			Body:     strings.Join(proseLines, "\n"),
			StartPos: proseStart,
			EndPos:   Position{Line: len(lines), Col: 1},
		})
	}

	return ParseResult{Blocks: blocks}
}

// splitLines splits source into lines without trailing newlines.
func splitLines(source string) []string {
	if source == "" {
		return nil
	}
	lines := strings.Split(source, "\n")
	// Remove trailing empty line from final newline.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// parseFenceOpening checks if a line is a fence opening. It returns the fence
// character (` or ~), the fence length, the indent, and the info string.
// Returns fenceLen=0 if the line is not a fence opening.
func parseFenceOpening(line string) (fenceChar byte, fenceLen int, indent int, info string) {
	// Count leading spaces (up to 3 for CommonMark).
	for indent < len(line) && indent < 3 && line[indent] == ' ' {
		indent++
	}

	if indent >= len(line) {
		return 0, 0, 0, ""
	}

	ch := line[indent]
	if ch != '`' && ch != '~' {
		return 0, 0, 0, ""
	}

	// Count fence characters.
	start := indent
	for fenceLen = 0; start+fenceLen < len(line) && line[start+fenceLen] == ch; fenceLen++ {
	}

	if fenceLen < 3 {
		return 0, 0, 0, ""
	}

	// Info string is everything after the fence characters.
	info = strings.TrimSpace(line[start+fenceLen:])

	// CommonMark: backtick fences cannot have backticks in the info string.
	if ch == '`' && strings.ContainsRune(info, '`') {
		return 0, 0, 0, ""
	}

	return ch, fenceLen, indent, info
}

// isClosingFence checks if a line is a valid closing fence for the given
// opening fence parameters.
func isClosingFence(line string, fenceChar byte, fenceLen int, maxIndent int) bool {
	// Count leading spaces.
	indent := 0
	for indent < len(line) && indent < 3 && line[indent] == ' ' {
		indent++
	}

	if indent >= len(line) {
		return false
	}

	if line[indent] != fenceChar {
		return false
	}

	// Count fence characters.
	count := 0
	for indent+count < len(line) && line[indent+count] == fenceChar {
		count++
	}

	if count < fenceLen {
		return false
	}

	// After the fence characters, only spaces are allowed (no info string).
	rest := strings.TrimSpace(line[indent+count:])
	return rest == ""
}

// classifyInfoString determines the block kind from the info string.
func classifyInfoString(info string) BlockKind {
	// Extract the first word (language identifier).
	lang := info
	if idx := strings.IndexAny(info, " \t"); idx >= 0 {
		lang = info[:idx]
	}

	switch lang {
	case "go:defs":
		return BlockDefs
	case "go:notebook":
		return BlockCell
	case "output":
		return BlockOutput
	default:
		// Plain "go" and everything else is prose.
		return BlockProse
	}
}

// removeIndent removes up to n spaces of indentation from a line.
func removeIndent(line string, n int) string {
	removed := 0
	for removed < n && removed < len(line) && line[removed] == ' ' {
		removed++
	}
	return line[removed:]
}
