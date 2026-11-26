# Charts

The chart system (`nb/chart`) renders data visualizations to SVG and PNG
directly from cell code. Charts appear as display output in the notebook.

## Chart Types

- **Line** (`chart.Line`) — polyline chart with X values and one or more
  named series
- **Area** (`chart.Area`) — filled polygon chart, optionally stacked
- **Bar** (`chart.Bar`) — grouped or stacked bar chart with category labels
- **Scatter** (`chart.Scatter`) — X/Y scatter plot with optional group
  coloring
- **Histogram** (`chart.Histogram`) — auto-binned histogram using Sturges'
  rule (bins override available)
- **Pie** (`chart.Pie`) — pie chart with category labels and proportional
  values
- **Heatmap** (`chart.Heatmap`) — color-intensity grid with row and column
  labels

## ctx.Chart Usage

```go
ctx.Chart(chart.Bar{
    Title:  "Monthly Revenue",
    Labels: []string{"Jan", "Feb", "Mar"},
    Series: []chart.Series{
        {Name: "2025", Values: []float64{120, 135, 148}},
    },
})
```

`ctx.Chart` calls `chart.RenderSVG` and stores the result as a
`image/svg+xml` display entry. The notebook server renders display entries
inline in the editor.

## SVG Rendering

Each `RenderXxxSVG` function produces a standalone SVG at a fixed viewport
(600×400 for most types, 400×400 for pie). Labels are HTML-escaped. The SVG
open tag embeds chart data as a JSON `data-chart` attribute for the
interactive script:

```html
<svg ... data-chart="{&quot;type&quot;:&quot;bar&quot;,...}">
```

## PNG Rendering

`chart.RenderPNG(ctx, eng, c, width, height)` converts the SVG to PNG via
the Typst WASM engine (`nb/engine`) and returns the result as a base64
string. This is used for PDF report output.

## Interactive Chart Script

The chart script is embedded in the server binary and served at
`/api/chart.js`. It is injected into rendered pages by the editor. The script
reads the `data-chart` attribute from each SVG, parses the JSON, and adds:

- **hover** — tooltip showing data point values
- **zoom** — drag to zoom into a region, double-click to reset
- **series toggle** — click a legend item to show or hide a series

## CSP Hash

The HTML render package (`nb/render/csp.go`) computes a SHA-256 hash of the
chart script for use in the Content-Security-Policy `script-src` directive.
Only the exact script content is permitted to execute; all other scripts
are blocked.
