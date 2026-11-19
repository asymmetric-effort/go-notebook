# Typst module license audit

## Scope

Every Rust crate compiled into the Typst WASM module must be under an
accepted license: MIT, Apache-2.0, or MPL-2.0. BSD is accepted only
where no alternative exists.

## Audit status

The full audit will be performed when the Typst module includes the
complete typst-as-lib compilation. The stub module's dependencies are:

| Crate | License | Status |
| --- | --- | --- |
| typst-library | Apache-2.0 | Accepted |
| comemo | MIT OR Apache-2.0 | Accepted |
| serde | MIT OR Apache-2.0 | Accepted |
| serde_json | MIT OR Apache-2.0 | Accepted |

## Full audit

Will list all transitive dependencies once the full module is built.
Run `cargo license` in `typst-module/` to generate the list.
