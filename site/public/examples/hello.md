# Hello World

This is a go-notebook. The prose sections are standard CommonMark Markdown
and appear as-is in the rendered output.

```go:defs
import "fmt"
import "runtime"
import "time"
```

## Greeting

```go:notebook name=greeting
fmt.Println("Hello from go-notebook!")
fmt.Printf("Built with Go %s on %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
```

## Current Time

```go:notebook name=timestamp
fmt.Printf("Report generated at: %s\n", time.Now().Format(time.RFC1123))
```

## Summary

This notebook demonstrated:

- Prose sections rendered as Markdown
- A `go:defs` block for shared imports
- Multiple `go:notebook` cells that produce output
- Named cells for clear error messages
