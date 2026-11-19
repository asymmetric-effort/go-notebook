# govulncheck exceptions

## Format

The `.govulncheck-exceptions` file in the repository root lists vulnerability
IDs that the project owner has reviewed and approved as exceptions.

Each non-empty, non-comment line must be a Go vulnerability ID in the form
`GO-YYYY-NNNN`. Lines starting with `#` are comments.

Example:

```text
# Approved by @sam-caldwell on 2026-11-19: does not affect our usage
GO-2024-1234
```

## Process

1. `govulncheck` runs in the pre-commit hook and in CI.
2. If it finds a vulnerability, the commit or build fails.
3. The project owner reviews the finding and either resolves it or adds
   it to `.govulncheck-exceptions` with a comment explaining why.
4. Invalid lines (not starting with `GO-`) produce a human-readable error.
