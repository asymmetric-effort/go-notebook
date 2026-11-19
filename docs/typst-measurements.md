# Typst engine measurements

## Methodology

Measurements are taken with Go benchmarks running the embedded Typst
WASM module on wazero.

## Results

### Module size

| Metric | Value |
| --- | --- |
| Stub WASM module | 40 KB |
| Full module (estimated) | 10-20 MB |

### Compile latency

To be measured with the full Typst module.

### Memory per compile

typst-go-wasm reports ~48 MB of linear memory per compile. To be
verified with our module.

### Binary size impact

The WASM module is embedded with `//go:embed`. Its size adds directly
to the report binary size.

## Running benchmarks

```bash
go test -bench=BenchmarkTypst -benchtime=5x ./nb/engine/
```
