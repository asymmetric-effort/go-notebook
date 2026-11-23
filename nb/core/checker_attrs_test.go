// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestCheckRefusedAttribute(t *testing.T) {
	src := "<div onclick=\"alert('xss')\">click me</div>\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "onclick") && strings.Contains(e.Error(), "rule 18") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 18 error for onclick, got: %v", errs)
	}
}

func TestCheckJavascriptURL(t *testing.T) {
	src := "<a href=\"javascript:alert('xss')\">link</a>\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "javascript:") && strings.Contains(e.Error(), "rule 18") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected rule 18 error for javascript: URL, got: %v", errs)
	}
}

func TestCheckAllowedAttributes(t *testing.T) {
	src := "<a href=\"https://go.dev\" class=\"link\">Go</a>\n"
	result := Parse(src)
	errs := Check(result)
	for _, e := range errs {
		if strings.Contains(e.Error(), "rule 18") {
			t.Errorf("allowed attributes should not error: %v", e)
		}
	}
}

func TestCheckOnMouseOver(t *testing.T) {
	src := "<span onmouseover=\"hack()\">hover</span>\n"
	result := Parse(src)
	errs := Check(result)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "onmouseover") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error for onmouseover, got: %v", errs)
	}
}
