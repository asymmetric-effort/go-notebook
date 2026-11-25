// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package commonmark

import (
	"strings"
	"testing"
)

func TestHeadings(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"# Heading 1", "<h1>Heading 1</h1>\n"},
		{"## Heading 2", "<h2>Heading 2</h2>\n"},
		{"### Heading 3", "<h3>Heading 3</h3>\n"},
		{"#### Heading 4", "<h4>Heading 4</h4>\n"},
		{"##### Heading 5", "<h5>Heading 5</h5>\n"},
		{"###### Heading 6", "<h6>Heading 6</h6>\n"},
		{"# Heading with trailing ###", "<h1>Heading with trailing</h1>\n"},
	}
	for _, tt := range tests {
		got := ToHTML(tt.input)
		if got != tt.want {
			t.Errorf("ToHTML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParagraphs(t *testing.T) {
	got := ToHTML("Hello world")
	if got != "<p>Hello world</p>\n" {
		t.Errorf("paragraph = %q", got)
	}
}

func TestMultipleParagraphs(t *testing.T) {
	got := ToHTML("First paragraph\n\nSecond paragraph")
	want := "<p>First paragraph</p>\n<p>Second paragraph</p>\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEmphasis(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"*emphasis*", "<p><em>emphasis</em></p>\n"},
		{"_emphasis_", "<p><em>emphasis</em></p>\n"},
		{"**strong**", "<p><strong>strong</strong></p>\n"},
		{"__strong__", "<p><strong>strong</strong></p>\n"},
	}
	for _, tt := range tests {
		got := ToHTML(tt.input)
		if got != tt.want {
			t.Errorf("ToHTML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCodeSpan(t *testing.T) {
	got := ToHTML("Use `fmt.Println` to print")
	want := "<p>Use <code>fmt.Println</code> to print</p>\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCodeSpanEscapes(t *testing.T) {
	got := ToHTML("Use `<div>` tag")
	if !strings.Contains(got, "&lt;div&gt;") {
		t.Errorf("code span should escape HTML: %q", got)
	}
}

func TestLinks(t *testing.T) {
	got := ToHTML("[Go](https://go.dev)")
	want := "<p><a href=\"https://go.dev\">Go</a></p>\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestImages(t *testing.T) {
	got := ToHTML("![alt text](image.png)")
	want := "<p><img src=\"image.png\" alt=\"alt text\" /></p>\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFencedCodeBlock(t *testing.T) {
	input := "```go\nfmt.Println(\"hello\")\n```"
	got := ToHTML(input)
	if !strings.Contains(got, "<pre><code class=\"language-go\">") {
		t.Errorf("missing language class: %q", got)
	}
	if !strings.Contains(got, "fmt.Println") {
		t.Errorf("missing code content: %q", got)
	}
}

func TestFencedCodeBlockNoLang(t *testing.T) {
	input := "```\nhello\n```"
	got := ToHTML(input)
	if !strings.Contains(got, "<pre><code>") {
		t.Errorf("missing code block: %q", got)
	}
}

func TestThematicBreak(t *testing.T) {
	for _, input := range []string{"---", "***", "___", "- - -", "* * *"} {
		got := ToHTML(input)
		if got != "<hr />\n" {
			t.Errorf("ToHTML(%q) = %q, want <hr />", input, got)
		}
	}
}

func TestUnorderedList(t *testing.T) {
	input := "- item 1\n- item 2\n- item 3"
	got := ToHTML(input)
	if !strings.Contains(got, "<ul>") {
		t.Errorf("missing <ul>: %q", got)
	}
	if !strings.Contains(got, "<li>item 1</li>") {
		t.Errorf("missing item 1: %q", got)
	}
	if strings.Count(got, "<li>") != 3 {
		t.Errorf("expected 3 items: %q", got)
	}
}

func TestOrderedList(t *testing.T) {
	input := "1. first\n2. second\n3. third"
	got := ToHTML(input)
	if !strings.Contains(got, "<ol>") {
		t.Errorf("missing <ol>: %q", got)
	}
	if !strings.Contains(got, "<li>first</li>") {
		t.Errorf("missing item: %q", got)
	}
}

func TestBlockquote(t *testing.T) {
	input := "> quoted text"
	got := ToHTML(input)
	if !strings.Contains(got, "<blockquote>") {
		t.Errorf("missing blockquote: %q", got)
	}
	if !strings.Contains(got, "quoted text") {
		t.Errorf("missing content: %q", got)
	}
}

func TestHTMLEscape(t *testing.T) {
	got := ToHTML("Use <div> & \"quotes\"")
	if !strings.Contains(got, "&lt;div&gt;") {
		t.Errorf("should escape <: %q", got)
	}
	if !strings.Contains(got, "&amp;") {
		t.Errorf("should escape &: %q", got)
	}
	if !strings.Contains(got, "&quot;quotes&quot;") {
		t.Errorf("should escape quotes: %q", got)
	}
}

func TestBackslashEscape(t *testing.T) {
	got := ToHTML(`\*not emphasis\*`)
	if strings.Contains(got, "<em>") {
		t.Errorf("backslash should escape: %q", got)
	}
}

func TestAutolink(t *testing.T) {
	got := ToHTML("<https://go.dev>")
	if !strings.Contains(got, "<a href=\"https://go.dev\">") {
		t.Errorf("autolink: %q", got)
	}
}

func TestEmpty(t *testing.T) {
	got := ToHTML("")
	if got != "" {
		t.Errorf("empty input = %q", got)
	}
}

func TestListWithStarMarker(t *testing.T) {
	got := ToHTML("* a\n* b")
	if !strings.Contains(got, "<ul>") || !strings.Contains(got, "<li>a</li>") {
		t.Errorf("star marker list: %q", got)
	}
}

func TestListWithPlusMarker(t *testing.T) {
	got := ToHTML("+ a\n+ b")
	if !strings.Contains(got, "<ul>") {
		t.Errorf("plus marker list: %q", got)
	}
}

func TestNestedBlockquote(t *testing.T) {
	got := ToHTML("> line1\n> line2")
	if !strings.Contains(got, "<blockquote>") {
		t.Errorf("blockquote: %q", got)
	}
}

func TestEmptyBlockquote(t *testing.T) {
	got := ToHTML(">")
	if !strings.Contains(got, "<blockquote>") {
		t.Errorf("empty blockquote: %q", got)
	}
}

func TestFencedCodeTilde(t *testing.T) {
	got := ToHTML("~~~\ncode\n~~~")
	if !strings.Contains(got, "<pre><code>") {
		t.Errorf("tilde fence: %q", got)
	}
}

func TestHTMLBlock(t *testing.T) {
	got := ToHTML("<div>\nhello\n</div>")
	if !strings.Contains(got, "<div>") {
		t.Errorf("html block: %q", got)
	}
}

func TestHTMLComment(t *testing.T) {
	got := ToHTML("<!-- comment -->")
	if !strings.Contains(got, "<!-- comment -->") {
		t.Errorf("html comment: %q", got)
	}
}

func TestHeadingEmpty(t *testing.T) {
	got := ToHTML("#")
	if !strings.Contains(got, "<h1></h1>") {
		t.Errorf("empty heading: %q", got)
	}
}

func TestNotHeading(t *testing.T) {
	got := ToHTML("#hashtag")
	if strings.Contains(got, "<h1>") {
		t.Errorf("should not be heading without space: %q", got)
	}
}

func TestSevenHashes(t *testing.T) {
	got := ToHTML("####### not a heading")
	if strings.Contains(got, "<h7>") {
		t.Errorf("7 hashes should not be heading: %q", got)
	}
}

func TestParagraphContinuation(t *testing.T) {
	got := ToHTML("line1\nline2")
	if !strings.Contains(got, "line1\nline2") {
		t.Errorf("continuation: %q", got)
	}
}

func TestListContinuation(t *testing.T) {
	got := ToHTML("- item1\n  continued")
	if !strings.Contains(got, "item1") && !strings.Contains(got, "continued") {
		t.Errorf("list continuation: %q", got)
	}
}

func TestOrderedListContinuation(t *testing.T) {
	got := ToHTML("1. item1\n   continued")
	if !strings.Contains(got, "item1") {
		t.Errorf("ordered continuation: %q", got)
	}
}

func TestBrokenLink(t *testing.T) {
	got := ToHTML("[text without close paren(url")
	if strings.Contains(got, "<a") {
		t.Errorf("broken link should not render: %q", got)
	}
}

func TestUnclosedCodeSpan(t *testing.T) {
	got := ToHTML("open `backtick")
	// Should render as literal backtick.
	if strings.Contains(got, "<code>") {
		t.Errorf("unclosed code span should not render: %q", got)
	}
}

func TestFencedCodeUnclosed(t *testing.T) {
	got := ToHTML("```\ncode without closing")
	if !strings.Contains(got, "code without closing") {
		t.Errorf("unclosed fence should still include content: %q", got)
	}
}

func TestThematicBreakNotEnough(t *testing.T) {
	got := ToHTML("--")
	if strings.Contains(got, "<hr") {
		t.Errorf("two dashes should not be HR: %q", got)
	}
}

func TestThematicBreakMixed(t *testing.T) {
	got := ToHTML("-*-")
	if strings.Contains(got, "<hr") {
		t.Errorf("mixed chars should not be HR: %q", got)
	}
}

func TestStrongEmphasisNested(t *testing.T) {
	got := ToHTML("***bold italic***")
	// Should render with strong (outer ** takes precedence).
	if !strings.Contains(got, "<strong>") {
		t.Errorf("nested emphasis should have strong: %q", got)
	}
}

func TestAutolinkNoScheme(t *testing.T) {
	got := ToHTML("<not a link>")
	if strings.Contains(got, "<a href") {
		t.Errorf("no scheme should not autolink: %q", got)
	}
}

func TestFenceCloserDifferentChar(t *testing.T) {
	got := ToHTML("```\ncode\n~~~")
	// ~~~ should not close a ``` fence.
	if !strings.Contains(got, "code") {
		t.Errorf("fence closer mismatch: %q", got)
	}
}

func TestDoubleBacktickCodeSpan(t *testing.T) {
	got := ToHTML("`` code `` with text")
	// Double backtick code spans are handled.
	if !strings.Contains(got, "<code>") {
		t.Errorf("double backtick should produce code: %q", got)
	}
}

func TestListHeadingInterrupt(t *testing.T) {
	// A list item on the first line of input should be a list, not a paragraph.
	got := ToHTML("- first")
	if !strings.Contains(got, "<ul>") {
		t.Errorf("first-line list: %q", got)
	}
}

func TestOrderedListFirstLine(t *testing.T) {
	got := ToHTML("1. first")
	if !strings.Contains(got, "<ol>") {
		t.Errorf("first-line ordered list: %q", got)
	}
}

func TestFenceCloserShorter(t *testing.T) {
	// Shorter fence should not close a longer fence.
	got := ToHTML("````\ncode\n```\n````")
	if !strings.Contains(got, "code") {
		t.Errorf("shorter fence closer: %q", got)
	}
}

func TestBrokenImage(t *testing.T) {
	got := ToHTML("![alt](")
	// Broken image syntax should not crash.
	if !strings.Contains(got, "![alt]") {
		t.Errorf("broken image: %q", got)
	}
}

func TestLinkMissingParen(t *testing.T) {
	got := ToHTML("[text]")
	// No (url) after ] — should not be a link.
	if strings.Contains(got, "<a") {
		t.Errorf("missing paren link: %q", got)
	}
}

func TestAutolinkSpace(t *testing.T) {
	got := ToHTML("<not a url with spaces>")
	if strings.Contains(got, "<a href") {
		t.Errorf("autolink with spaces: %q", got)
	}
}

func TestStripListMarkerNoMatch(t *testing.T) {
	// Directly test stripListMarker with non-matching input.
	result := stripListMarker("not a list item")
	if result != "not a list item" {
		t.Errorf("stripListMarker = %q", result)
	}
}

func TestStripOrderedMarkerNoMatch(t *testing.T) {
	result := stripOrderedMarker("not ordered")
	if result != "not ordered" {
		t.Errorf("stripOrderedMarker = %q", result)
	}
}

func TestParagraphInterruptedByHeading(t *testing.T) {
	got := ToHTML("text\n# heading")
	if !strings.Contains(got, "<p>text</p>") || !strings.Contains(got, "<h1>heading</h1>") {
		t.Errorf("heading interrupt: %q", got)
	}
}

func TestParagraphInterruptedByHR(t *testing.T) {
	got := ToHTML("text\n---")
	if !strings.Contains(got, "<p>text</p>") || !strings.Contains(got, "<hr />") {
		t.Errorf("HR interrupt: %q", got)
	}
}

func TestParagraphInterruptedByFence(t *testing.T) {
	got := ToHTML("text\n```\ncode\n```")
	if !strings.Contains(got, "<p>text</p>") || !strings.Contains(got, "<pre><code>") {
		t.Errorf("fence interrupt: %q", got)
	}
}

func TestParagraphInterruptedByBlockquote(t *testing.T) {
	got := ToHTML("text\n> quote")
	if !strings.Contains(got, "<p>text</p>") || !strings.Contains(got, "<blockquote>") {
		t.Errorf("blockquote interrupt: %q", got)
	}
}

func TestParagraphInterruptedByList(t *testing.T) {
	got := ToHTML("text\n- item")
	if !strings.Contains(got, "<p>text</p>") || !strings.Contains(got, "<ul>") {
		t.Errorf("list interrupt: %q", got)
	}
}

func TestParagraphInterruptedByOrderedList(t *testing.T) {
	got := ToHTML("text\n1. item")
	if !strings.Contains(got, "<p>text</p>") || !strings.Contains(got, "<ol>") {
		t.Errorf("ordered list interrupt: %q", got)
	}
}

func TestNewlineInAutolink(t *testing.T) {
	got := ToHTML("<broken\nlink>")
	// Newline in angle brackets should not be an autolink.
	if strings.Contains(got, "<a href") {
		t.Errorf("newline autolink: %q", got)
	}
}

func TestCodeSpanWithBacktickInside(t *testing.T) {
	got := ToHTML("`` ` ``")
	if !strings.Contains(got, "<code>") {
		t.Errorf("backtick inside code span: %q", got)
	}
}

func TestImageMissingCloseParen(t *testing.T) {
	got := ToHTML("![alt](url")
	// Should not render as image.
	if strings.Contains(got, "<img") {
		t.Errorf("missing close paren image: %q", got)
	}
}

func TestMixedContent(t *testing.T) {
	input := "# Title\n\nSome **bold** and *italic* text.\n\n- item 1\n- item 2\n\n---\n\n```go\ncode\n```\n"
	got := ToHTML(input)
	if !strings.Contains(got, "<h1>Title</h1>") {
		t.Error("missing heading")
	}
	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Error("missing strong")
	}
	if !strings.Contains(got, "<em>italic</em>") {
		t.Error("missing emphasis")
	}
	if !strings.Contains(got, "<ul>") {
		t.Error("missing list")
	}
	if !strings.Contains(got, "<hr />") {
		t.Error("missing thematic break")
	}
	if !strings.Contains(got, "<pre><code") {
		t.Error("missing code block")
	}
}

func TestBlockquoteFollowedByNonBlockquote(t *testing.T) {
	// A blockquote line followed by a non-blockquote line triggers the break.
	src := "> quoted\nnormal paragraph"
	got := ToHTML(src)
	if !strings.Contains(got, "<blockquote>") {
		t.Error("should have blockquote")
	}
	if !strings.Contains(got, "normal paragraph") {
		t.Error("should have paragraph after blockquote")
	}
}

func TestHTMLBlockBreakOnBlankLine(t *testing.T) {
	// An HTML block ending with a blank line triggers the break.
	src := "<div class=\"test\">\nsome content\n\nmore content"
	got := ToHTML(src)
	if !strings.Contains(got, "<div") {
		t.Error("should contain HTML block")
	}
}

func TestCodeSpanStripsSpaces(t *testing.T) {
	// A code span with spaces on both ends: ` code ` -> strips outer spaces.
	got := ToHTML("text ` code ` more")
	if !strings.Contains(got, "<code>code</code>") {
		t.Errorf("code span should strip outer spaces: %q", got)
	}
}

func TestFindAutolinkNoClosingAngle(t *testing.T) {
	// findAutolink with no closing > returns -1 (falls through to end).
	src := "text <notclosed"
	got := ToHTML(src)
	// Should not crash; autolink won't be inserted.
	_ = got
}

func TestOrderedListContinuationLine(t *testing.T) {
	// An ordered list item followed by a non-list, non-blank line (continuation path).
	src := "1. first item\n   continued on next line\n2. second item"
	got := ToHTML(src)
	if !strings.Contains(got, "<ol>") {
		t.Errorf("should have ordered list: %q", got)
	}
	if !strings.Contains(got, "continued") {
		t.Errorf("continuation text should be included: %q", got)
	}
}
