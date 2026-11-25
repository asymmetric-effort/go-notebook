# CommonMark features in Typst/PDF output

This document lists CommonMark features and their support status in the
Typst markup output used for PDF compilation.

## Supported features

- ATX headings (H1-H6)
- Paragraphs
- Emphasis (`*em*`) — partial, regex-based
- Strong (`**strong**`)
- Code spans (`` `code` ``) — via `#raw`
- Fenced code blocks
- Indented code blocks — via `#raw`
- Unordered lists
- Ordered lists — via `+` items
- Links — via `#link`
- Autolinks — via `#link`
- Images — via `#image`
- Blockquotes — via `#quote`
- Thematic breaks — via `#line`
- Tables (GFM extension) — via `#table`
- Hard line breaks — via `\`

## Unsupported or limited features

- **Raw HTML blocks** — intentionally unsupported (HTML cannot be
  meaningfully converted to Typst)
- **Inline HTML** — tags like `<br>`, `<sup>`, `<sub>` not handled
- **Setext headings** — only ATX headings (`#`) are converted
- **Link reference definitions** — only inline links are converted
- **Strikethrough (GFM)** — `~~text~~` not converted
- **Task lists (GFM)** — `- [ ]` / `- [x]` not converted
- **Footnotes** — not part of CommonMark
- **Nested emphasis** — `***bold italic***` limited

## Design decisions

- **Raw HTML is intentionally unsupported in PDF.** HTML is a fundamentally
  different rendering model from Typst. The checker allows only allowlisted
  HTML tags in prose (rule 18), and these are designed for the HTML backend.
  PDF output focuses on the Markdown/Typst subset.

- **GFM extensions** (strikethrough, task lists) are not part of the
  CommonMark specification and are lower priority.
