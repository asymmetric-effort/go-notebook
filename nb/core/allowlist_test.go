// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import "testing"

func TestIsAllowedTag(t *testing.T) {
	allowed := []string{"a", "b", "div", "span", "table", "img", "style", "details"}
	for _, tag := range allowed {
		if !IsAllowedTag(tag) {
			t.Errorf("IsAllowedTag(%q) = false, want true", tag)
		}
	}
}

func TestIsRefusedTag(t *testing.T) {
	refused := []string{"script", "iframe", "frame", "object", "embed", "base", "meta"}
	for _, tag := range refused {
		if !IsRefusedTag(tag) {
			t.Errorf("IsRefusedTag(%q) = false, want true", tag)
		}
	}
}

func TestIsAllowedTagUnknown(t *testing.T) {
	if IsAllowedTag("custom-element") {
		t.Error("unknown tags should not be allowed")
	}
}

func TestIsRefusedAttribute(t *testing.T) {
	refused := []string{"onclick", "onload", "onerror", "onmouseover", "onfocus", "onCustom"}
	for _, attr := range refused {
		if !IsRefusedAttribute(attr) {
			t.Errorf("IsRefusedAttribute(%q) = false, want true", attr)
		}
	}
}

func TestIsRefusedAttributeAllowed(t *testing.T) {
	allowed := []string{"class", "id", "href", "src", "alt", "style"}
	for _, attr := range allowed {
		if IsRefusedAttribute(attr) {
			t.Errorf("IsRefusedAttribute(%q) = true, want false", attr)
		}
	}
}

func TestIsJavascriptURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"javascript:alert(1)", true},
		{"JavaScript:void(0)", true},
		{"  javascript:test", true},
		{"https://example.com", false},
		{"#anchor", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsJavascriptURL(tt.url); got != tt.want {
			t.Errorf("IsJavascriptURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

func TestIsAllowedTagCaseInsensitive(t *testing.T) {
	if !IsAllowedTag("DIV") {
		t.Error("tag check should be case-insensitive")
	}
}

func TestIsRefusedTagCaseInsensitive(t *testing.T) {
	if !IsRefusedTag("SCRIPT") {
		t.Error("refused tag check should be case-insensitive")
	}
}
