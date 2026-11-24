```go:defs
import "github.com/asymmetric-effort/go-notebook/nb/chart"
```

```go:notebook name=hostile exported(done bool)
ctx.Chart(chart.Bar{
	Title: "<script>alert('xss')</script>",
	Labels: []string{"<img onerror=alert(1)>", "normal"},
	Series: []chart.Series{{Name: "<b>bold</b>", Values: []float64{10, 20}}},
})
done = true
```
