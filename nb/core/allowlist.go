// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import "strings"

// AllowedTags is the allowlist of HTML tags permitted in notebook prose.
// Unknown tags are refused by default (allowlist, not blocklist).
var AllowedTags = map[string]bool{
	"a": true, "b": true, "i": true, "em": true, "strong": true,
	"sub": true, "sup": true, "br": true, "hr": true,
	"span": true, "div": true, "p": true,
	"table": true, "thead": true, "tbody": true, "tr": true, "th": true, "td": true,
	"ul": true, "ol": true, "li": true,
	"dl": true, "dt": true, "dd": true,
	"details": true, "summary": true,
	"blockquote": true, "pre": true, "code": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"img": true, "style": true, "link": true,
}

// RefusedTags are explicitly refused with an error.
var RefusedTags = map[string]bool{
	"script": true, "iframe": true, "frame": true,
	"object": true, "embed": true, "base": true,
	"meta": true,
}

// RefusedAttributes are event handlers and dangerous attributes.
var RefusedAttributes = map[string]bool{
	"onload": true, "onerror": true, "onclick": true, "onmouseover": true,
	"onfocus": true, "onblur": true, "onsubmit": true, "onchange": true,
	"onkeydown": true, "onkeyup": true, "onkeypress": true,
}

// IsAllowedTag checks if an HTML tag name is on the allowlist.
func IsAllowedTag(tag string) bool {
	return AllowedTags[strings.ToLower(tag)]
}

// IsRefusedTag checks if an HTML tag is explicitly refused.
func IsRefusedTag(tag string) bool {
	return RefusedTags[strings.ToLower(tag)]
}

// IsRefusedAttribute checks if an attribute name is a refused event handler.
func IsRefusedAttribute(attr string) bool {
	lower := strings.ToLower(attr)
	if RefusedAttributes[lower] {
		return true
	}
	// Refuse all on* attributes.
	return strings.HasPrefix(lower, "on")
}

// IsJavascriptURL checks if a URL uses a dangerous scheme (javascript:,
// vbscript:, data:text/html). These can execute code when loaded.
func IsJavascriptURL(url string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(url))
	switch {
	case strings.HasPrefix(trimmed, "javascript:"):
		return true
	case strings.HasPrefix(trimmed, "vbscript:"):
		return true
	case strings.HasPrefix(trimmed, "data:text/html"):
		return true
	default:
		return false
	}
}
