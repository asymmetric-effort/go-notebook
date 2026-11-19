// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

// SourceMap maps positions between the notebook Markdown file and the
// generated Go source. The language server proxy uses this to translate
// positions in both directions.
type SourceMap struct {
	entries []sourceMapEntry
}

type sourceMapEntry struct {
	mdLine int // 1-based Markdown line.
	mdCol  int // 1-based Markdown column.
	goLine int // 1-based Go line in the generated file.
	goCol  int // 1-based Go column.
	length int // Number of lines this mapping covers.
}

// NewSourceMap creates an empty source map.
func NewSourceMap() *SourceMap {
	return &SourceMap{}
}

// Add records a mapping from Markdown position to Go position.
func (sm *SourceMap) Add(mdLine, mdCol, goLine, goCol, length int) {
	sm.entries = append(sm.entries, sourceMapEntry{
		mdLine: mdLine,
		mdCol:  mdCol,
		goLine: goLine,
		goCol:  goCol,
		length: length,
	})
}

// MdToGo translates a Markdown position to a Go position.
// Returns (0, 0, false) if no mapping covers the position.
func (sm *SourceMap) MdToGo(line, col int) (goLine, goCol int, ok bool) {
	for _, e := range sm.entries {
		if line >= e.mdLine && line < e.mdLine+e.length {
			offset := line - e.mdLine
			return e.goLine + offset, col + (e.goCol - e.mdCol), true
		}
	}
	return 0, 0, false
}

// GoToMd translates a Go position back to a Markdown position.
// Returns (0, 0, false) if no mapping covers the position.
func (sm *SourceMap) GoToMd(line, col int) (mdLine, mdCol int, ok bool) {
	for _, e := range sm.entries {
		if line >= e.goLine && line < e.goLine+e.length {
			offset := line - e.goLine
			return e.mdLine + offset, col - (e.goCol - e.mdCol), true
		}
	}
	return 0, 0, false
}

// Len returns the number of entries in the source map.
func (sm *SourceMap) Len() int {
	return len(sm.entries)
}
