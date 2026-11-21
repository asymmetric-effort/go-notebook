# Compile cost per cell

## Measurement

The compile cost of `go run` per cell is measured with Go benchmarks
in `nb/kernel/bench_test.go`. Each benchmark runs `go run` on a
generated cell package to measure the actual cost a notebook author
pays for each ad hoc run.

## Running the benchmarks

```bash
go test -bench=BenchmarkGoRun -benchtime=5x -timeout 120s ./nb/kernel/
```

## Results

Measured on Intel Core i7-7700HQ @ 2.80GHz, Go 1.27.4, Linux amd64:

| Scenario | Time per cell | Notes |
| --- | --- | --- |
| Warm cache, simple cell | ~56ms | Build cache hit, link step only |
| Warm cache, with imports | ~55ms | fmt + strings, cache warm |
| Cold cache, simple cell | ~1-2s (estimated) | Full compile, varies by deps |

## Analysis

- **56ms per cell with warm cache** is fast enough for interactive use.
  An author clicking "run" in the editor will see results in under 100ms
  for cells that haven't changed.
- The Go build cache is effective: warm-cache runs are dominated by the
  link step, not compilation.
- Cold-cache runs (first run after editing a cell) will be 1-2 seconds
  for simple cells, longer for cells with many transitive dependencies.
- cgo cells will be slower due to C compiler invocation.

## Factors

- **Build cache warm:** subsequent runs of an unchanged cell use the Go
  build cache, so only the link step runs.
- **Build cache cold:** first run after changing a cell or its imports
  requires a full compile.
- **cgo cells:** cells that import cgo packages pay the C compiler cost.
- **Dependencies:** cells with many transitive imports compile slower.
