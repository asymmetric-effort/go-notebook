# yamlfmt lint mode verification

Verified on 2026-11-19 with yamlfmt v0.21.0.

`yamlfmt -lint <file>` exits non-zero when the file is not formatted
and does not modify the file. This satisfies the requirement that
linters never modify files.

Decision: use yamlfmt with the `-lint` flag for YAML linting in `make lint`.
No custom Go YAML linter is needed.
