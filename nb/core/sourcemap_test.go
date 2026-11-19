// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import "testing"

func TestSourceMapRoundTrip(t *testing.T) {
	sm := NewSourceMap()
	// Map Markdown lines 10-14 to Go lines 20-24.
	sm.Add(10, 1, 20, 1, 5)

	// Forward: Markdown -> Go.
	goLine, goCol, ok := sm.MdToGo(12, 5)
	if !ok {
		t.Fatal("MdToGo returned not ok")
	}
	if goLine != 22 {
		t.Errorf("goLine = %d, want 22", goLine)
	}
	if goCol != 5 {
		t.Errorf("goCol = %d, want 5", goCol)
	}

	// Reverse: Go -> Markdown.
	mdLine, mdCol, ok := sm.GoToMd(22, 5)
	if !ok {
		t.Fatal("GoToMd returned not ok")
	}
	if mdLine != 12 {
		t.Errorf("mdLine = %d, want 12", mdLine)
	}
	if mdCol != 5 {
		t.Errorf("mdCol = %d, want 5", mdCol)
	}
}

func TestSourceMapNotFound(t *testing.T) {
	sm := NewSourceMap()
	sm.Add(10, 1, 20, 1, 5)

	_, _, ok := sm.MdToGo(5, 1) // Before the mapping.
	if ok {
		t.Error("MdToGo should not find line 5")
	}

	_, _, ok = sm.GoToMd(30, 1) // After the mapping.
	if ok {
		t.Error("GoToMd should not find line 30")
	}
}

func TestSourceMapMultipleEntries(t *testing.T) {
	sm := NewSourceMap()
	sm.Add(5, 1, 10, 1, 3)  // Lines 5-7 -> 10-12
	sm.Add(20, 1, 30, 1, 2) // Lines 20-21 -> 30-31

	goLine, _, ok := sm.MdToGo(6, 1)
	if !ok || goLine != 11 {
		t.Errorf("MdToGo(6) = %d, %v; want 11, true", goLine, ok)
	}

	goLine, _, ok = sm.MdToGo(21, 1)
	if !ok || goLine != 31 {
		t.Errorf("MdToGo(21) = %d, %v; want 31, true", goLine, ok)
	}
}

func TestSourceMapEmpty(t *testing.T) {
	sm := NewSourceMap()
	if sm.Len() != 0 {
		t.Errorf("Len() = %d, want 0", sm.Len())
	}
	_, _, ok := sm.MdToGo(1, 1)
	if ok {
		t.Error("empty source map should not find any position")
	}
}

func TestSourceMapLen(t *testing.T) {
	sm := NewSourceMap()
	sm.Add(1, 1, 1, 1, 1)
	sm.Add(2, 1, 2, 1, 1)
	if sm.Len() != 2 {
		t.Errorf("Len() = %d, want 2", sm.Len())
	}
}

func TestSourceMapColumnOffset(t *testing.T) {
	sm := NewSourceMap()
	// Markdown col 5 maps to Go col 1 (indent removed).
	sm.Add(10, 5, 20, 1, 3)

	goLine, goCol, ok := sm.MdToGo(11, 8)
	if !ok {
		t.Fatal("MdToGo returned not ok")
	}
	if goLine != 21 {
		t.Errorf("goLine = %d, want 21", goLine)
	}
	if goCol != 4 { // 8 - (5-1) = 4
		t.Errorf("goCol = %d, want 4", goCol)
	}

	mdLine, mdCol, ok := sm.GoToMd(21, 4)
	if !ok {
		t.Fatal("GoToMd returned not ok")
	}
	if mdLine != 11 {
		t.Errorf("mdLine = %d, want 11", mdLine)
	}
	if mdCol != 8 {
		t.Errorf("mdCol = %d, want 8", mdCol)
	}
}
