# Compile cost per cell

## Measurement

The compile cost of `go run` per cell is measured with the Go benchmark
framework. Each benchmark runs `go run` on an extracted cell package to
measure the actual cost a notebook author pays for each ad hoc run.

## Running the benchmark

```bash
go test -bench=BenchmarkCompileCost -benchtime=5x ./nb/kernel/
```

## Expected results

Results will be recorded here after M1 is complete and example notebooks
exist. The Go build cache should keep reruns of unchanged cells fast;
cgo cells will be slower.

## Factors

- **Build cache warm:** subsequent runs of an unchanged cell use the Go
  build cache, so only the link step runs.
- **Build cache cold:** first run after changing a cell or its imports
  requires a full compile.
- **cgo cells:** cells that import cgo packages pay the C compiler cost.
- **Dependencies:** cells with many transitive imports compile slower.
