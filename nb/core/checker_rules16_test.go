// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package core

import (
	"strings"
	"testing"
)

func TestCheckRule16AbsoluteImagePath(t *testing.T) {
	src := "![logo](/etc/passwd)\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for absolute image path (rule 16)")
	}
	if !strings.Contains(errs[0].Error(), "rule 16") {
		t.Errorf("error = %q", errs[0].Error())
	}
}

func TestCheckRule16ParentTraversal(t *testing.T) {
	src := "![logo](../secret/image.png)\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for ../ in image path (rule 16)")
	}
}

func TestCheckRule16ValidPath(t *testing.T) {
	src := "![logo](images/logo.png)\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("valid image path should not produce errors: %v", errs)
	}
}

func TestCheckRule16WebURLAllowed(t *testing.T) {
	src := "![chart](https://example.com/chart.png)\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("web URL should be allowed: %v", errs)
	}
}

func TestCheckRule18ScriptTag(t *testing.T) {
	src := "<script>alert('xss')</script>\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for <script> (rule 18)")
	}
	if !strings.Contains(errs[0].Error(), "rule 18") {
		t.Errorf("error = %q", errs[0].Error())
	}
}

func TestCheckRule18IframeTag(t *testing.T) {
	src := "<iframe src=\"https://evil.com\"></iframe>\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for <iframe> (rule 18)")
	}
}

func TestCheckRule18AllowedTags(t *testing.T) {
	src := "<b>bold</b> <i>italic</i> <div>block</div>\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("allowed tags should not produce errors: %v", errs)
	}
}

func TestCheckRule19Doctype(t *testing.T) {
	src := "<!DOCTYPE html>\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for <!DOCTYPE> (rule 19)")
	}
	if !strings.Contains(errs[0].Error(), "rule 19") {
		t.Errorf("error = %q", errs[0].Error())
	}
}

func TestCheckRule19Entity(t *testing.T) {
	src := "<!ENTITY xxe SYSTEM \"file:///etc/passwd\">\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) == 0 {
		t.Fatal("expected error for <!ENTITY> (rule 19)")
	}
}

func TestCheckRule18ObjectEmbed(t *testing.T) {
	src := "<object data=\"bad.swf\"></object>\n<embed src=\"bad.swf\">\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) < 2 {
		t.Errorf("expected 2 errors for object+embed, got %d", len(errs))
	}
}

func TestCheckCleanProse(t *testing.T) {
	src := "# Title\n\nSome **bold** text with an ![image](logo.png).\n\n<details><summary>More</summary>Content</details>\n"
	result := Parse(src)
	errs := Check(result)
	if len(errs) != 0 {
		t.Errorf("clean prose should have no errors: %v", errs)
	}
}
