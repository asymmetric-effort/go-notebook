// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"fmt"
	"strings"
)

// WriteOutputBlocks replaces or inserts output blocks in the Markdown source
// after each cell's closing fence. If an output block already exists for a
// cell, it is replaced. If not, a new one is inserted.
func WriteOutputBlocks(source string, outputs map[string]string) string {
	if len(outputs) == 0 {
		return source
	}

	result := Parse(source)
	lines := strings.Split(source, "\n")

	// Process in reverse order to preserve line numbers.
	for i := len(result.Blocks) - 1; i >= 0; i-- {
		b := result.Blocks[i]
		if b.Kind != BlockCell {
			continue
		}
		attrs, err := ParseInfoString(b.Info, b.StartPos.Line)
		if err != nil || attrs.Name == "" {
			continue
		}

		output, ok := outputs[attrs.Name]
		if !ok {
			continue
		}

		// Find or create the output block position.
		insertLine := b.EndPos.Line // 1-based, the closing fence line.

		// Check if the next non-empty block is an output block for this cell.
		existingOutputStart, existingOutputEnd := findExistingOutput(result.Blocks, i, attrs.Name)

		outputBlock := formatOutputBlock(attrs.Name, output)

		if existingOutputStart > 0 {
			// Replace existing output block.
			newLines := make([]string, 0, len(lines))
			newLines = append(newLines, lines[:existingOutputStart-1]...)
			newLines = append(newLines, outputBlock...)
			newLines = append(newLines, lines[existingOutputEnd:]...)
			lines = newLines
		} else {
			// Insert after closing fence.
			newLines := make([]string, 0, len(lines)+len(outputBlock)+1)
			newLines = append(newLines, lines[:insertLine]...)
			newLines = append(newLines, "")
			newLines = append(newLines, outputBlock...)
			newLines = append(newLines, lines[insertLine:]...)
			lines = newLines
		}
	}

	return strings.Join(lines, "\n")
}

// findExistingOutput looks for an output block following a cell block.
func findExistingOutput(blocks []Block, cellIdx int, cellName string) (startLine, endLine int) {
	for j := cellIdx + 1; j < len(blocks); j++ {
		b := blocks[j]
		if b.Kind == BlockProse {
			// Skip whitespace-only prose between cell and output.
			if strings.TrimSpace(b.Body) == "" {
				continue
			}
			return 0, 0
		}
		if b.Kind == BlockOutput {
			// Check if it's for this cell.
			info := b.Info
			if strings.Contains(info, "name="+cellName) {
				return b.StartPos.Line, b.EndPos.Line
			}
			return 0, 0
		}
		return 0, 0
	}
	return 0, 0
}

// formatOutputBlock creates the lines for an output fenced block.
func formatOutputBlock(name, content string) []string {
	return []string{
		fmt.Sprintf("```output name=%s", name),
		content,
		"```",
	}
}
