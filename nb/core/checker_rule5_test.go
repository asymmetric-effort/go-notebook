// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import "testing"

func TestCheckRule5ValidReference(t *testing.T) {
	// Cell b references cell a which is above it — should not error.
	src := "```go:notebook name=a exported(val int)\nval = 1\n```\n\n" +
		"```go:notebook name=b exported(result int)\nresult = a.val + 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("valid upstream reference should not error: %v", errs)
	}
}

func TestCheckRule5ContextReference(t *testing.T) {
	// ctx.Open should not be flagged.
	src := "```go:notebook name=test reads(file.txt:embed)\nf, _ := ctx.Open(\"file.txt\")\n_ = f\n```\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("ctx references should not be flagged: %v", errs)
	}
}
