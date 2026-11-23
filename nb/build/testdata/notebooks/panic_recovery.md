# Panic Recovery

A cell that panics should be recovered, producing a partial report.

Some prose before the error.

```go:notebook name=broken exported(val int)
panic("intentional test panic")
```
