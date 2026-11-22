# CommonMark features unsupported in Typst/PDF output

This document lists CommonMark features that render correctly in HTML
but are not yet supported or have limited support in the Typst markup
output (`--format typst`).

## Supported features

| Feature | HTML | Typst |
| --- | --- | --- |
| ATX headings (H1-H6) | Yes | Yes |
| Paragraphs | Yes | Yes |
| Emphasis (`*em*`) | Yes | Partial (regex-based) |
| Strong (`**strong**`) | Yes | Yes |
| Code spans (`` `code` ``) | Yes | Yes (`#raw`) |
| Fenced code blocks | Yes | Yes |
| Unordered lists | Yes | Yes |
| Ordered lists | Yes | Yes (`+` items) |
| Links | Yes | Yes (`#link`) |
| Images | Yes | Yes (`#image`) |
| Blockquotes | Yes | Yes (`#quote`) |
| Thematic breaks | Yes | Yes (`#line`) |

## Unsupported or limited features

| Feature | HTML | Typst | Notes |
| --- | --- | --- | --- |
| Raw HTML blocks | Yes | No | HTML tags are not converted to Typst equivalents |
| Inline HTML | Yes | No | Tags like `<br>`, `<sup>`, `<sub>` are not handled |
| Autolinks | Yes | No | `<https://url>` not converted to `#link` |
| Hard line breaks | Yes | No | Trailing spaces or `\` not converted |
| Setext headings | Yes | No | Only ATX headings (`#`) are converted |
| Indented code blocks | Yes | No | Only fenced code blocks are converted |
| Link reference definitions | Yes | No | Only inline links are converted |
| Tables (GFM extension) | Yes | No | Tables not converted to Typst `#table` |
| Strikethrough (GFM) | Yes | No | `~~text~~` not converted |
| Task lists (GFM) | Yes | No | `- [ ]` / `- [x]` not converted |
| Footnotes | Yes | No | Not part of CommonMark |
| Nested emphasis | Partial | No | `***bold italic***` limited |

## Plan

These gaps will be addressed incrementally as the Typst markup
conversion matures. The M2 Typst WASM engine will enable PDF
compilation and testing, after which visual regression tests will
catch rendering differences automatically.
