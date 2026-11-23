// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import "testing"

// FuzzAllowlistTag fuzzes the tag allowlist with random tag names.
// It must never panic.
func FuzzAllowlistTag(f *testing.F) {
	f.Add("script")
	f.Add("SCRIPT")
	f.Add("ScRiPt")
	f.Add("iframe")
	f.Add("div")
	f.Add("a")
	f.Add("")
	f.Add("x-custom")
	f.Add("on")
	f.Add("<script>")
	f.Add("script/src=evil")

	f.Fuzz(func(t *testing.T, tag string) {
		_ = IsAllowedTag(tag)
		_ = IsRefusedTag(tag)
	})
}

// FuzzAllowlistAttribute fuzzes the attribute allowlist.
func FuzzAllowlistAttribute(f *testing.F) {
	f.Add("onclick")
	f.Add("ONCLICK")
	f.Add("onload")
	f.Add("href")
	f.Add("class")
	f.Add("")
	f.Add("onfoo")
	f.Add("on")
	f.Add("style")

	f.Fuzz(func(t *testing.T, attr string) {
		_ = IsRefusedAttribute(attr)
	})
}

// FuzzJavascriptURL fuzzes the javascript: URL detection.
func FuzzJavascriptURL(f *testing.F) {
	f.Add("javascript:alert(1)")
	f.Add("JAVASCRIPT:alert(1)")
	f.Add("  javascript:alert(1)")
	f.Add("https://example.com")
	f.Add("")
	f.Add("java\tscript:alert(1)")
	f.Add("data:text/html,<script>alert(1)</script>")

	f.Fuzz(func(t *testing.T, url string) {
		_ = IsJavascriptURL(url)
	})
}
