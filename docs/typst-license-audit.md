# Typst module license audit

## Scope

Every Rust crate compiled into the Typst WASM module must be under an
accepted license. This audit covers all transitive dependencies.

## Accepted licenses

- MIT
- Apache-2.0
- BSD-2-Clause, BSD-3-Clause
- MPL-2.0
- Unicode-3.0
- Zlib, 0BSD

## Audit results

Generated with `cargo license` in `typst-module/` on 2026-11-22.

### By license

| License | Count | Examples |
| --- | --- | --- |
| Apache-2.0 OR MIT | 157 | serde, syn, rand, rayon, image, pdf-writer |
| Apache-2.0 | 18 | typst, typst-library, typst-pdf, typst-eval |
| MIT | 43 | bincode, fontdb, rustybuzz, syntect, typst-module |
| (Apache-2.0 OR MIT) AND Unicode-3.0 | 3 | unicode-ident, unicode-normalization |
| Unicode-3.0 | 38 | icu_collections, icu_properties, zerovec |
| Apache-2.0 OR MIT OR Zlib | 9 | bytemuck, miniz_oxide, tinyvec |
| BSD-2-Clause | 3 | arrayref, kamadak-exif, mutate_once |
| BSD-3-Clause | 2 | tiny-skia, tiny-skia-path |
| MPL-2.0 | 2 | resvg, usvg |
| Zlib | 3 | foldhash, slotmap, zlib-rs |
| MIT OR Unlicense | 8 | aho-corasick, memchr, csv |
| Other permissive | 5 | 0BSD, BSL-1.0, Apache WITH LLVM-exception |

### Status

**All ~290 transitive dependencies use permissive licenses.** No GPL,
AGPL, LGPL, or other copyleft licenses found.

### Notable crates

| Crate | License | Role |
| --- | --- | --- |
| typst | Apache-2.0 | Compiler core |
| typst-library | Apache-2.0 | Standard library and World trait |
| typst-pdf | Apache-2.0 | PDF generation |
| wazero (Go side) | Apache-2.0 | WASM runtime (not in Rust audit) |
| resvg | MPL-2.0 | SVG rendering |
| usvg | MPL-2.0 | SVG parsing |

## Re-auditing

Run after any dependency update:

```bash
cd typst-module && cargo license
```
