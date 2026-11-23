# Multi-Cell Report

Two cells where the second reads from the first.

```go:notebook name=load exported(val int)
val = 10
```

The load cell produced a value.

```go:notebook name=process exported(result int)
result = load.val * 3
```
