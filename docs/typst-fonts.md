# Embedded fonts

## Current status

The Typst WASM module renders PDFs using Typst's built-in font
fallbacks. No external fonts are embedded in the module. Text renders
correctly for ASCII content; non-Latin scripts may fall back to
placeholder glyphs.

## Font selection

Fonts must be under MIT, Apache-2.0, or OFL (SIL Open Font License)
which permits embedding in distributed binaries.

### Selected fonts

| Font | License | Purpose | Status |
| --- | --- | --- | --- |
| New Computer Modern | OFL 1.1 | Serif body text (Typst default) | Planned |
| Inter | OFL 1.1 | Sans-serif headings | Planned |
| JetBrains Mono | OFL 1.1 | Monospace code | Planned |

### License audit

All three fonts use the SIL Open Font License 1.1 which explicitly
permits:

- Embedding in documents and applications
- Redistribution in source and binary forms
- Use in commercial and non-commercial contexts

The OFL requires that fonts not be sold by themselves but permits
bundling with other software (our use case).

## Implementation plan

Fonts will be embedded via one of these approaches:

1. **In the WASM module**: Include font files in the Rust binary and
   register them with the Typst `World::font()` trait method. Increases
   WASM module size by ~5-10 MB.

2. **In the Go runtime**: Embed font files with `//go:embed` in the
   `nb/engine` package and pass them to the WASM module via stdin along
   with the compilation command.

3. **Via typst-assets**: The `typst-assets` crate (already a transitive
   dependency) may provide access to Typst's default font set.

Approach 3 is preferred as it leverages existing infrastructure.

## Running without fonts

The current module works for all ASCII text and common Unicode.
PDFs are valid but may display Typst's fallback font warnings for
content requiring specific typefaces.
