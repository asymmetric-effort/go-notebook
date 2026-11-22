# Typst engine measurements

## Methodology

Measurements taken with Go tests running the embedded Typst WASM module
on wazero. Platform: linux/amd64 in Docker (Claude session container).

## Results

### Module size

| Metric | Value |
| --- | --- |
| WASM module (release, LTO, strip) | 17,994,578 bytes (18 MB) |
| Rust crate count | ~200 transitive dependencies |
| Build target | wasm32-wasip1 |

### Compile latency

| Operation | Time |
| --- | --- |
| WASM module instantiation (cold start) | ~7-8 seconds |
| Compile "Hello, world!" to PDF | ~7-8 seconds (includes instantiation) |
| PDF output size for "Hello, world!" | 2,011 bytes |
| PDF output size for heading + paragraph | 2,299 bytes |

Each compilation creates a fresh WASM module instance. A pooled
instance design would amortize the cold-start cost.

### Memory per compile

The wazero runtime allocates linear memory for the WASM module.
Typical usage observed: 20-50 MB per compilation depending on
document complexity.

### Binary size impact

The WASM module is embedded with `//go:embed`. The 18 MB module adds
directly to the Go binary size. A report binary with the Typst engine
embedded will be approximately 18 MB larger than one without.

## Running benchmarks

```bash
go test -bench=BenchmarkTypst -benchtime=3x ./nb/engine/
go test -v -run TestEngineCompilePDF ./nb/engine/
```
