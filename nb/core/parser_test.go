// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"testing"
)

func TestParseEmpty(t *testing.T) {
	result := Parse("")
	if len(result.Blocks) != 0 {
		t.Errorf("Parse(\"\") = %d blocks, want 0", len(result.Blocks))
	}
}

func TestParseProse(t *testing.T) {
	result := Parse("# Hello\n\nSome text.\n")
	if len(result.Blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(result.Blocks))
	}
	if result.Blocks[0].Kind != BlockProse {
		t.Errorf("kind = %v, want prose", result.Blocks[0].Kind)
	}
}

func TestParseDefs(t *testing.T) {
	src := "# Title\n\n```go:defs\nimport \"fmt\"\n```\n"
	result := Parse(src)

	var defs []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockDefs {
			defs = append(defs, b)
		}
	}
	if len(defs) != 1 {
		t.Fatalf("got %d defs blocks, want 1", len(defs))
	}
	if defs[0].Body != "import \"fmt\"" {
		t.Errorf("body = %q", defs[0].Body)
	}
	if defs[0].Info != "go:defs" {
		t.Errorf("info = %q, want %q", defs[0].Info, "go:defs")
	}
}

func TestParseCell(t *testing.T) {
	src := "```go:notebook name=load exported(rows []Row)\nf, err := ctx.Open(\"sales.csv\")\n```\n"
	result := Parse(src)

	var cells []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cells = append(cells, b)
		}
	}
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	if cells[0].Info != "go:notebook name=load exported(rows []Row)" {
		t.Errorf("info = %q", cells[0].Info)
	}
	if cells[0].Body != "f, err := ctx.Open(\"sales.csv\")" {
		t.Errorf("body = %q", cells[0].Body)
	}
}

func TestParseOutput(t *testing.T) {
	src := "```output name=totals\n- East: 120400.00\n```\n"
	result := Parse(src)

	var outputs []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockOutput {
			outputs = append(outputs, b)
		}
	}
	if len(outputs) != 1 {
		t.Fatalf("got %d output blocks, want 1", len(outputs))
	}
	if outputs[0].Body != "- East: 120400.00" {
		t.Errorf("body = %q", outputs[0].Body)
	}
}

func TestParsePlainGoIsProse(t *testing.T) {
	src := "```go\nfmt.Println(\"hello\")\n```\n"
	result := Parse(src)

	for _, b := range result.Blocks {
		if b.Kind == BlockCell || b.Kind == BlockDefs {
			t.Errorf("plain go block classified as %v, want prose", b.Kind)
		}
	}
}

func TestParseDocumentOrder(t *testing.T) {
	src := "# Title\n\n```go:defs\nimport \"fmt\"\n```\n\nSome prose.\n\n```go:notebook name=main\nfmt.Println(\"hi\")\n```\n\n```output name=main\nhi\n```\n"
	result := Parse(src)

	var kinds []BlockKind
	for _, b := range result.Blocks {
		kinds = append(kinds, b.Kind)
	}

	if len(kinds) < 4 {
		t.Fatalf("got %d blocks, want at least 4", len(kinds))
	}

	// Verify defs comes before cell, cell before output.
	defsIdx, cellIdx, outputIdx := -1, -1, -1
	for i, k := range kinds {
		if k == BlockDefs && defsIdx == -1 {
			defsIdx = i
		}
		if k == BlockCell && cellIdx == -1 {
			cellIdx = i
		}
		if k == BlockOutput && outputIdx == -1 {
			outputIdx = i
		}
	}
	if defsIdx >= cellIdx {
		t.Errorf("defs at %d, cell at %d — defs should come first", defsIdx, cellIdx)
	}
	if cellIdx >= outputIdx {
		t.Errorf("cell at %d, output at %d — cell should come first", cellIdx, outputIdx)
	}
}

func TestParsePositions(t *testing.T) {
	src := "line1\n```go:notebook name=test\nbody line\n```\nline5\n"
	result := Parse(src)

	var cell Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cell = b
			break
		}
	}

	if cell.StartPos.Line != 2 {
		t.Errorf("StartPos.Line = %d, want 2", cell.StartPos.Line)
	}
	if cell.BodyPos.Line != 3 {
		t.Errorf("BodyPos.Line = %d, want 3", cell.BodyPos.Line)
	}
	if cell.EndPos.Line != 4 {
		t.Errorf("EndPos.Line = %d, want 4", cell.EndPos.Line)
	}
}

func TestParseTildeFence(t *testing.T) {
	src := "~~~go:notebook name=test\nbody\n~~~\n"
	result := Parse(src)

	var cells []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cells = append(cells, b)
		}
	}
	if len(cells) != 1 {
		t.Fatalf("got %d cells from tilde fence, want 1", len(cells))
	}
}

func TestParseLongerFenceContainsShorter(t *testing.T) {
	src := "````go:notebook name=test\n```\nstill body\n```\n````\n"
	result := Parse(src)

	var cells []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cells = append(cells, b)
		}
	}
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	if cells[0].Body != "```\nstill body\n```" {
		t.Errorf("body = %q, want inner fences preserved", cells[0].Body)
	}
}

func TestParseIndentedFence(t *testing.T) {
	// Up to 3 spaces of indentation is allowed for fenced code blocks.
	src := "   ```go:notebook name=test\n   body\n   ```\n"
	result := Parse(src)

	var cells []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cells = append(cells, b)
		}
	}
	if len(cells) != 1 {
		t.Fatalf("got %d cells from indented fence, want 1", len(cells))
	}
	if cells[0].Body != "body" {
		t.Errorf("body = %q, want %q (indent should be removed)", cells[0].Body, "body")
	}
}

func TestParseFourSpaceIndentNotFence(t *testing.T) {
	// 4 spaces is a code block in CommonMark, not a fenced block.
	src := "    ```go:notebook name=test\n    body\n    ```\n"
	result := Parse(src)

	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			t.Error("4-space indented fence should not be parsed as a cell")
		}
	}
}

func TestParseUnclosedFence(t *testing.T) {
	src := "```go:notebook name=test\nbody line\n"
	result := Parse(src)

	var cells []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cells = append(cells, b)
		}
	}
	if len(cells) != 1 {
		t.Fatalf("got %d cells from unclosed fence, want 1", len(cells))
	}
	if cells[0].Body != "body line" {
		t.Errorf("body = %q", cells[0].Body)
	}
}

func TestParseBacktickInTildeInfoString(t *testing.T) {
	// Tilde fences can have backticks in the info string.
	src := "~~~go:notebook name=`test`\nbody\n~~~\n"
	result := Parse(src)

	var cells []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cells = append(cells, b)
		}
	}
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1 (backtick in tilde info string is valid)", len(cells))
	}
}

func TestParseBacktickInBacktickInfoString(t *testing.T) {
	// Backtick fences cannot have backticks in the info string.
	src := "```go:notebook name=`test`\nbody\n```\n"
	result := Parse(src)

	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			t.Error("backtick in backtick info string should not parse as a fence")
		}
	}
}

func TestParseMultipleCells(t *testing.T) {
	src := "```go:notebook name=a\nbody a\n```\n\n```go:notebook name=b\nbody b\n```\n"
	result := Parse(src)

	var cells []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cells = append(cells, b)
		}
	}
	if len(cells) != 2 {
		t.Fatalf("got %d cells, want 2", len(cells))
	}
	if cells[0].Body != "body a" {
		t.Errorf("first cell body = %q", cells[0].Body)
	}
	if cells[1].Body != "body b" {
		t.Errorf("second cell body = %q", cells[1].Body)
	}
}

func TestParseEmptyBody(t *testing.T) {
	src := "```go:notebook name=empty\n```\n"
	result := Parse(src)

	var cells []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cells = append(cells, b)
		}
	}
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	if cells[0].Body != "" {
		t.Errorf("body = %q, want empty", cells[0].Body)
	}
}

func TestParseClosingFenceMustMatchChar(t *testing.T) {
	// Tilde opening cannot be closed by backticks.
	src := "~~~go:notebook name=test\nbody\n```\n~~~\n"
	result := Parse(src)

	var cells []Block
	for _, b := range result.Blocks {
		if b.Kind == BlockCell {
			cells = append(cells, b)
		}
	}
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	// The ``` line should be part of the body, not a closing fence.
	if cells[0].Body != "body\n```" {
		t.Errorf("body = %q, want backtick line preserved in body", cells[0].Body)
	}
}

func TestBlockKindString(t *testing.T) {
	tests := []struct {
		kind BlockKind
		want string
	}{
		{BlockProse, "prose"},
		{BlockDefs, "defs"},
		{BlockCell, "cell"},
		{BlockOutput, "output"},
		{BlockKind(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestClassifyInfoString(t *testing.T) {
	tests := []struct {
		info string
		want BlockKind
	}{
		{"go:defs", BlockDefs},
		{"go:notebook name=test", BlockCell},
		{"go:notebook", BlockCell},
		{"output name=test", BlockOutput},
		{"output", BlockOutput},
		{"go", BlockProse},
		{"python", BlockProse},
		{"", BlockProse},
		{"javascript", BlockProse},
	}
	for _, tt := range tests {
		if got := classifyInfoString(tt.info); got != tt.want {
			t.Errorf("classifyInfoString(%q) = %v, want %v", tt.info, got, tt.want)
		}
	}
}

func TestRemoveIndent(t *testing.T) {
	tests := []struct {
		line string
		n    int
		want string
	}{
		{"   hello", 3, "hello"},
		{"   hello", 5, "hello"},
		{"hello", 3, "hello"},
		{"  hello", 3, "hello"},
		{"", 3, ""},
	}
	for _, tt := range tests {
		if got := removeIndent(tt.line, tt.n); got != tt.want {
			t.Errorf("removeIndent(%q, %d) = %q, want %q", tt.line, tt.n, got, tt.want)
		}
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a\nb\n", 2},
		{"a\nb", 2},
		{"a\n", 1},
	}
	for _, tt := range tests {
		got := splitLines(tt.input)
		if len(got) != tt.want {
			t.Errorf("splitLines(%q) = %d lines, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestParseFenceOpening(t *testing.T) {
	tests := []struct {
		line     string
		wantLen  int
		wantInfo string
	}{
		{"```go:notebook", 3, "go:notebook"},
		{"````go:defs", 4, "go:defs"},
		{"~~~output name=test", 3, "output name=test"},
		{"   ```go:notebook", 3, "go:notebook"},
		{"``not a fence", 0, ""},
		{"normal text", 0, ""},
		{"    ```too much indent", 0, ""},
	}
	for _, tt := range tests {
		_, fenceLen, _, info := parseFenceOpening(tt.line)
		if fenceLen != tt.wantLen {
			t.Errorf("parseFenceOpening(%q): fenceLen = %d, want %d", tt.line, fenceLen, tt.wantLen)
		}
		if fenceLen > 0 && info != tt.wantInfo {
			t.Errorf("parseFenceOpening(%q): info = %q, want %q", tt.line, info, tt.wantInfo)
		}
	}
}

func TestIsClosingFence(t *testing.T) {
	tests := []struct {
		line     string
		char     byte
		fenceLen int
		indent   int
		want     bool
	}{
		{"```", '`', 3, 0, true},
		{"````", '`', 3, 0, true},
		{"```  ", '`', 3, 0, true},
		{"``", '`', 3, 0, false},
		{"~~~", '`', 3, 0, false},
		{"~~~", '~', 3, 0, true},
		{"```info", '`', 3, 0, false},
		{"   ```", '`', 3, 0, true},
	}
	for _, tt := range tests {
		got := isClosingFence(tt.line, tt.char, tt.fenceLen, tt.indent)
		if got != tt.want {
			t.Errorf("isClosingFence(%q, %c, %d, %d) = %v, want %v",
				tt.line, tt.char, tt.fenceLen, tt.indent, got, tt.want)
		}
	}
}
