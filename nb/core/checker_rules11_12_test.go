// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestCheckRule11SecretDeclared(t *testing.T) {
	// ctx.Secret with secret declared in secrets() — no error.
	src := "```go:notebook name=test secrets(db) exported(val int)\nconn := ctx.Secret(\"db\")\n_ = conn\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 11") {
			t.Errorf("declared secret should not error: %v", e)
		}
	}
}

func TestCheckRule11SecretUndeclared(t *testing.T) {
	// ctx.Secret without declaring in secrets() — should error.
	src := "```go:notebook name=test exported(val int)\nconn := ctx.Secret(\"db\")\n_ = conn\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 11") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 11 error for undeclared secret, got: %v", errs)
	}
}

func TestCheckRule12SecretToExport(t *testing.T) {
	// Direct assignment of ctx.Secret to exported field — should error.
	src := "```go:notebook name=test secrets(db) exported(password string)\npassword = ctx.Secret(\"db\")\n```\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 12") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 12 error for secret-to-export, got: %v", errs)
	}
}

func TestCheckRule12SecretToLocal(t *testing.T) {
	// Secret assigned to a local variable — no error.
	src := "```go:notebook name=test secrets(db) exported(val int)\nconn := ctx.Secret(\"db\")\n_ = conn\nval = 1\n```\n"
	result := Parse(src)
	errs := Check(result)
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 12") {
			t.Errorf("secret to local should not trigger rule 12: %v", e)
		}
	}
}

func TestExportedFieldNames(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"val int, msg string", []string{"val", "msg"}},
		{"val int", []string{"val"}},
		{"", nil},
	}
	for _, tt := range tests {
		fields := exportedFieldNames(tt.input)
		for _, w := range tt.want {
			if !fields[w] {
				t.Errorf("exportedFieldNames(%q) missing %q", tt.input, w)
			}
		}
	}
}
