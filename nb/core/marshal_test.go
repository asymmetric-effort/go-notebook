// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestGenerateMarshal(t *testing.T) {
	code := GenerateMarshal("loadCell", []string{"rows []Row", "total float64"})
	if !strings.Contains(code, "func (c loadCell) MarshalBinary()") {
		t.Error("missing MarshalBinary method")
	}
	if !strings.Contains(code, "func (c *loadCell) UnmarshalBinary(data []byte)") {
		t.Error("missing UnmarshalBinary method")
	}
	if !strings.Contains(code, "enc.Encode(c.rows)") {
		t.Error("missing rows encode")
	}
	if !strings.Contains(code, "dec.Decode(&c.total)") {
		t.Error("missing total decode")
	}
}

func TestGenerateMarshalEmpty(t *testing.T) {
	code := GenerateMarshal("emptyCell", nil)
	if !strings.Contains(code, "MarshalBinary") {
		t.Error("missing MarshalBinary for empty struct")
	}
}

func TestDeepCopyCode(t *testing.T) {
	code := DeepCopyCode("loadCell")
	if !strings.Contains(code, "func deepCopyLoadCell(src loadCell) loadCell") {
		t.Error("missing deep copy function")
	}
	if !strings.Contains(code, "MarshalBinary()") {
		t.Error("missing MarshalBinary call")
	}
	if !strings.Contains(code, "UnmarshalBinary(data)") {
		t.Error("missing UnmarshalBinary call")
	}
}

func TestFieldName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"rows []Row", "rows"},
		{"total float64", "total"},
		{"m map[string]int", "m"},
		{"x", "x"},
	}
	for _, tt := range tests {
		if got := fieldName(tt.input); got != tt.want {
			t.Errorf("fieldName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
