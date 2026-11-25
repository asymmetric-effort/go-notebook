// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestCheckRule1ExportedRequiresName(t *testing.T) {
	src := "```go:notebook exported(val int)\nx := 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for exported without name (rule 1)")
	}
	if !strings.Contains(errs[0].Error(), "rule 1") {
		t.Errorf("error = %q, want rule 1", errs[0].Error())
	}
}

func TestCheckRule1NamedExportedPasses(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
}

func TestCheckRule2DefsOnlyAllowed(t *testing.T) {
	// var without //go:embed should fail.
	src := "```go:defs\nvar x int = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for var in go:defs (rule 14)")
	}
	if !strings.Contains(errs[0].Error(), "rule 14") {
		t.Errorf("error = %q, want rule 14", errs[0].Error())
	}
}

func TestCheckRule2DefsAllowedDecls(t *testing.T) {
	src := "```go:defs\nimport \"fmt\"\n\nconst x = 1\n\ntype Row struct{ Name string }\n\nfunc parse() {}\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("unexpected errors for allowed defs: %v", errs)
	}
}

func TestCheckRule3NoInit(t *testing.T) {
	src := "```go:defs\nfunc init() {}\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for func init() (rule 3)")
	}
	if !strings.Contains(errs[0].Error(), "rule 3") {
		t.Errorf("error = %q, want rule 3", errs[0].Error())
	}
}

func TestCheckRule10DirectOsOpen(t *testing.T) {
	src := "```go:notebook name=test\nf, err := os.Open(\"file.txt\")\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for direct os.Open (rule 10)")
	}
	if !strings.Contains(errs[0].Error(), "rule 10") {
		t.Errorf("error = %q, want rule 10", errs[0].Error())
	}
}

func TestCheckRule10DirectOsReadFile(t *testing.T) {
	src := "```go:notebook name=test\ndata, err := os.ReadFile(\"file.txt\")\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for direct os.ReadFile (rule 10)")
	}
}

func TestCheckRule10CtxOpenPasses(t *testing.T) {
	src := "```go:notebook name=test reads(file.txt:embed)\nf, err := ctx.Open(\"file.txt\")\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("ctx.Open should be allowed: %v", errs)
	}
}

func TestCheckRule15EmbedInCell(t *testing.T) {
	src := "```go:notebook name=test\n//go:embed file.txt\nvar data string\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for //go:embed in cell (rule 15)")
	}
	if !strings.Contains(errs[0].Error(), "rule 15") {
		t.Errorf("error = %q, want rule 15", errs[0].Error())
	}
}

func TestCheckRule15GenerateInCell(t *testing.T) {
	src := "```go:notebook name=test\n//go:generate echo hello\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for //go:generate in cell (rule 15)")
	}
}

func TestCheckCleanNotebook(t *testing.T) {
	src := "```go:defs\nimport \"fmt\"\n\nconst currency = \"USD\"\n```\n\n```go:notebook name=load exported(val int)\nval = 42\n```\n\n```go:notebook\nfmt.Println(load.val)\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("clean notebook should have no errors: %v", errs)
	}
}

func TestCheckEmptyNotebook(t *testing.T) {
	result := Parse("")
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("empty notebook should have no errors: %v", errs)
	}
}

func TestCheckRule14VarWithEmbed(t *testing.T) {
	src := "```go:defs\n//go:embed file.txt\nvar data string\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("var with //go:embed should be allowed: %v", errs)
	}
}

func TestCheckDefsUnparseable(t *testing.T) {
	// Invalid Go in defs should not cause checker to crash.
	src := "```go:defs\nthis is not valid go\n```\n"
	result := Parse(src)
	errs := Check(result)
	// No errors from checker — the Go compiler will catch it.
	if len(errs) != 0 {
		t.Errorf("unparseable defs should not produce checker errors: %v", errs)
	}
}

func TestCheckInvalidInfoString(t *testing.T) {
	src := "```go:notebook name=\nbody\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for invalid info string")
	}
}

func TestCheckFieldNameEmpty(t *testing.T) {
	if got := fieldName(""); got != "" {
		t.Errorf("fieldName(\"\") = %q, want empty", got)
	}
}

func TestCheckErrorLine(t *testing.T) {
	src := "# Title\n\n```go:notebook exported(val int)\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error")
	}
	ce, ok := errs[0].(*CheckError)
	if !ok {
		t.Fatalf("error type = %T, want *CheckError", errs[0])
	}
	if ce.Line != 3 {
		t.Errorf("Line = %d, want 3", ce.Line)
	}
}

func TestCheckRule7ChanType(t *testing.T) {
	src := "```go:notebook name=test exported(ch chan int)\n_ = ch\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 7") && strings.Contains(e.Error(), "chan") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 7 error for chan type, got: %v", errs)
	}
}

func TestCheckRule7FuncType(t *testing.T) {
	src := "```go:notebook name=test exported(fn func())\n_ = fn\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 7") && strings.Contains(e.Error(), "func") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 7 error for func type, got: %v", errs)
	}
}

func TestCheckRule7UnsafePointer(t *testing.T) {
	src := "```go:notebook name=test exported(p unsafe.Pointer)\n_ = p\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 7") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 7 error for unsafe.Pointer, got: %v", errs)
	}
}

func TestCheckRule7ReceiveChan(t *testing.T) {
	src := "```go:notebook name=test exported(ch <-chan int)\n_ = ch\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 7") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 7 error for <-chan type, got: %v", errs)
	}
}

func TestCheckRule7SerializableTypesPass(t *testing.T) {
	types := []string{
		"val int",
		"msg string",
		"rows []float64",
		"m map[string]int",
		"data []byte",
	}
	for _, field := range types {
		src := "```go:notebook name=test exported(" + field + ")\n_ = val\n```\n"
		result := Parse(src)
		errs := Check(result)
		for _, e := range errs {
			if strings.Contains(e.Error(), "rule 7") {
				t.Errorf("type %q should be serializable, got rule 7 error: %v", field, e)
			}
		}
	}
}
