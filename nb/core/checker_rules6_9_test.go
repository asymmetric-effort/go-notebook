// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestCheckRule6AssignField(t *testing.T) {
	// Cell b assigns to cell a's field — should error.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\na.val = 99\nresult = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 6") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 6 error for assigning to a.val, got: %v", errs)
	}
}

func TestCheckRule6TakeAddress(t *testing.T) {
	// Taking address of another cell's field — should error.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\np := &a.val\n_ = p\nresult = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 6") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 6 error for &a.val, got: %v", errs)
	}
}

func TestCheckRule6ReadOK(t *testing.T) {
	// Reading another cell's field is fine — no error.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\nresult = a.val\n```\n"
	result := Parse(src)
	errs := Check(result)
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 6") {
			t.Errorf("reading a.val should not trigger rule 6: %v", e)
		}
	}
}

func TestCheckRule8OpenDeclared(t *testing.T) {
	// ctx.Open with file declared in reads — no error.
	src := "```go:notebook name=test reads(data.csv:embed)\nf, _ := ctx.Open(\"data.csv\")\n_ = f\n```\n"
	result := Parse(src)
	errs := Check(result)
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 8") {
			t.Errorf("declared read should not error: %v", e)
		}
	}
}

func TestCheckRule8OpenUndeclared(t *testing.T) {
	// ctx.Open without declaring in reads — should error.
	src := "```go:notebook name=test exported(val int)\nf, _ := ctx.Open(\"data.csv\")\n_ = f\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 8") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 8 error for undeclared read, got: %v", errs)
	}
}

func TestCheckRule9CreateDeclared(t *testing.T) {
	// ctx.Create with file declared in writes — no error.
	src := "```go:notebook name=test writes(output.txt)\nw, _ := ctx.Create(\"output.txt\")\n_ = w\n```\n"
	result := Parse(src)
	errs := Check(result)
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 9") {
			t.Errorf("declared write should not error: %v", e)
		}
	}
}

func TestCheckRule9CreateUndeclared(t *testing.T) {
	// ctx.Create without declaring in writes — should error.
	src := "```go:notebook name=test exported(val int)\nw, _ := ctx.Create(\"output.txt\")\n_ = w\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 9") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 9 error for undeclared write, got: %v", errs)
	}
}

func TestCheckRule6NoUpstreamCells(t *testing.T) {
	// First cell with no upstream — no rule 6 check needed.
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 6") {
			t.Errorf("first cell should not trigger rule 6: %v", e)
		}
	}
}
