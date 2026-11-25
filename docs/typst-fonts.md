# Embedded fonts

## Current status

The Typst WASM module embeds fonts from the `typst-assets` crate
(version 0.13, `fonts` feature). Fonts are loaded at module
initialization via `typst_assets::fonts()` and registered with the
Typst font book. The module is ~27 MB, with fonts accounting for
the majority of the size.

## Embedded font families

The `typst-assets` crate provides Typst's default font set:

| Font                | License                        | Purpose                           |
| ------------------- | ------------------------------ | --------------------------------- |
| New Computer Modern | OFL 1.1                        | Serif body text (Typst default)   |
| Libertinus          | OFL 1.1                        | Math and serif fallback           |
| DejaVu              | Bitstream Vera + public domain | Sans-serif and monospace fallback |

These fonts cover Latin, Cyrillic, Greek, and common Unicode symbols.
CJK and other complex scripts may still use system fallbacks if
available, or Typst's placeholder glyphs if not.

## License audit

All embedded fonts use licenses that permit embedding and redistribution:

- **OFL 1.1** (New Computer Modern, Libertinus): permits embedding in
  documents and applications, redistribution in source and binary forms,
  commercial and non-commercial use. Requires fonts not be sold standalone.
- **Bitstream Vera / public domain** (DejaVu): no restrictions on
  redistribution or embedding.

The full license audit is in `docs/typst-license-audit.md`.

## Implementation

Fonts are embedded via approach 3 (typst-assets crate) in
`typst-module/Cargo.toml`:

```toml
typst-assets = { version = "0.13", features = ["fonts"] }
```

The Rust code loads them at initialization (`typst-module/src/main.rs`):

```rust
for font_data in typst_assets::fonts() {
    let buffer = Bytes::new(font_data.to_vec());
    for font in Font::iter(buffer) {
        book.push(font.info().clone());
        fonts.push(font);
    }
}
```

## Binary size impact

The `--no-pdf` build flag excludes the Typst WASM module (~27 MB)
from the report binary when PDF output is not needed.
