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
	// Issue #314: empty fields must not declare enc/dec variables.
	if strings.Contains(code, "enc :=") || strings.Contains(code, "dec :=") {
		t.Errorf("empty fields should not declare enc/dec variables:\n%s", code)
	}
	if !strings.Contains(code, "return nil, nil") {
		t.Error("MarshalBinary should return nil, nil for empty fields")
	}
	if !strings.Contains(code, "return nil") {
		t.Error("UnmarshalBinary should return nil for empty fields")
	}
}

func TestGenerateMarshalOnlyErrField(t *testing.T) {
	// A single "err" field is filtered out, leaving zero marshalable fields.
	// Issue #314: must not generate unused enc/dec variables.
	code := GenerateMarshal("errOnlyCell", []string{"err error"})
	if strings.Contains(code, "enc :=") || strings.Contains(code, "dec :=") {
		t.Errorf("err-only fields should not declare enc/dec variables:\n%s", code)
	}
	if !strings.Contains(code, "return nil, nil") {
		t.Error("MarshalBinary should return nil, nil when only field is err")
	}
}

func TestGenerateMarshalEmptyNoBannedDecls(t *testing.T) {
	// Verify that the generated code for empty fields does not contain
	// any variable declarations that would cause "declared and not used"
	// compile errors.
	code := GenerateMarshal("noFieldCell", nil)
	for _, banned := range []string{"enc :=", "dec :=", "var buf"} {
		if strings.Contains(code, banned) {
			t.Errorf("empty-field output must not contain %q to avoid compile errors:\n%s", banned, code)
		}
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

func TestGenerateMarshalSkipsErrField(t *testing.T) {
	// The err field is an interface (error) that gob can't encode when nil.
	// GenerateMarshal must skip it.
	code := GenerateMarshal("dataCell", []string{"err error", "val int", "msg string"})
	if strings.Contains(code, "enc.Encode(c.err)") {
		t.Error("marshal should skip err field")
	}
	if strings.Contains(code, "dec.Decode(&c.err)") {
		t.Error("unmarshal should skip err field")
	}
	if !strings.Contains(code, "enc.Encode(c.val)") {
		t.Error("missing val encode")
	}
	if !strings.Contains(code, "enc.Encode(c.msg)") {
		t.Error("missing msg encode")
	}
}

func TestExtractStructFields(t *testing.T) {
	code := "type testCell struct {\n\terr error\n\tval int\n\tmsg string\n}"
	fields := extractStructFields(code)
	if len(fields) != 3 {
		t.Fatalf("got %d fields, want 3", len(fields))
	}
	if fields[0] != "err error" {
		t.Errorf("fields[0] = %q", fields[0])
	}
	if fields[1] != "val int" {
		t.Errorf("fields[1] = %q", fields[1])
	}
	if fields[2] != "msg string" {
		t.Errorf("fields[2] = %q", fields[2])
	}
}

func TestExtractStructFieldsEmpty(t *testing.T) {
	code := "type emptyCell struct {\n}"
	fields := extractStructFields(code)
	if len(fields) != 0 {
		t.Errorf("got %d fields for empty struct", len(fields))
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
