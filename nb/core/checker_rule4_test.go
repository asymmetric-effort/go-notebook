// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestCheckRule4BareReturnAllowed(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nval = 42\nreturn\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("bare return should be allowed: %v", errs)
	}
}

func TestCheckRule4ReturnValueRejected(t *testing.T) {
	src := "```go:notebook name=test exported(val int)\nreturn 42\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 4") {
			found = true
		}
	}
	if !found {
		t.Error("expected rule 4 error for return <value>")
	}
}

func TestCheckRule4ReturnMultipleRejected(t *testing.T) {
	src := "```go:notebook name=test exported(a int)\nreturn 1, nil\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 4") {
			found = true
		}
	}
	if !found {
		t.Error("expected rule 4 error for return with multiple values")
	}
}

func TestCheckRule4NoReturnAllowed(t *testing.T) {
	// A cell body with no return at all — the deferred function handles it.
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("no return should be fine: %v", errs)
	}
}

func TestCheckRule4InvalidGoIgnored(t *testing.T) {
	// Invalid Go in cell body should not cause checker to panic.
	src := "```go:notebook name=test\nthis is not go code {{\n```\n"
	result := Parse(src)
	errs := Check(result)
	// Should not panic, may or may not have errors.
	_ = errs
}
