# Charts

go-notebook includes a built-in chart rendering engine. Charts are created
using the `ctx.Chart()` API inside notebook cells and render as SVG in HTML
output or inline in PDF.

## Supported Chart Types

| Type | Description |
|------|-------------|
| `line` | Line chart with multiple series |
| `bar` | Vertical bar chart |
| `scatter` | Scatter plot with X/Y data points |
| `pie` | Pie chart with labeled slices |
| `histogram` | Frequency distribution histogram |
| `heatmap` | 2D grid with color intensity |
| `box` | Box-and-whisker plot for statistical distributions |

## Basic Usage

```go
ctx.Chart(chart.Options{
    Type:   "line",
    Title:  "Monthly Revenue",
    Width:  600,
    Height: 400,
    XLabel: "Month",
    YLabel: "Revenue ($)",
    Series: []chart.Series{
        {
            Name:   "2025",
            Values: []float64{10, 25, 30, 45, 50, 65},
            Labels: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"},
        },
    },
})
```

## Line Chart

Line charts display data points connected by straight lines. Use them for
time series or continuous data.

```go
ctx.Chart(chart.Options{
    Type:  "line",
    Title: "Temperature Over Time",
    Series: []chart.Series{
        {Name: "Sensor A", Values: []float64{20, 22, 19, 25, 28, 24}},
        {Name: "Sensor B", Values: []float64{18, 20, 21, 23, 26, 22}},
    },
})
```

## Bar Chart

Bar charts compare discrete categories.

```go
ctx.Chart(chart.Options{
    Type:  "bar",
    Title: "Sales by Region",
    Series: []chart.Series{
        {
            Name:   "Q1",
            Values: []float64{150, 230, 180, 310},
            Labels: []string{"North", "South", "East", "West"},
        },
    },
})
```

## Scatter Plot

Scatter plots show the relationship between two variables.

```go
ctx.Chart(chart.Options{
    Type:  "scatter",
    Title: "Height vs Weight",
    Series: []chart.Series{
        {
            Name:   "Sample",
            XY:     [][2]float64{{160, 55}, {170, 70}, {175, 75}, {180, 85}},
        },
    },
})
```

## Pie Chart

Pie charts show proportional composition.

```go
ctx.Chart(chart.Options{
    Type:  "pie",
    Title: "Market Share",
    Series: []chart.Series{
        {
            Values: []float64{45, 25, 20, 10},
            Labels: []string{"Product A", "Product B", "Product C", "Other"},
        },
    },
})
```

## Histogram

Histograms show frequency distributions of continuous data.

```go
ctx.Chart(chart.Options{
    Type:  "histogram",
    Title: "Response Times",
    Series: []chart.Series{
        {Values: []float64{12, 15, 14, 18, 22, 19, 25, 30, 28, 35}},
    },
    Bins: 5,
})
```

## Heatmap

Heatmaps display 2D data using color intensity.

```go
ctx.Chart(chart.Options{
    Type:  "heatmap",
    Title: "Correlation Matrix",
    Grid:  [][]float64{
        {1.0, 0.8, 0.3},
        {0.8, 1.0, 0.5},
        {0.3, 0.5, 1.0},
    },
    XLabels: []string{"A", "B", "C"},
    YLabels: []string{"A", "B", "C"},
})
```

## Box Plot

Box plots summarize statistical distributions showing median, quartiles,
and outliers.

```go
ctx.Chart(chart.Options{
    Type:  "box",
    Title: "Test Scores by Class",
    Series: []chart.Series{
        {Name: "Class A", Values: []float64{72, 85, 90, 88, 76, 95, 68}},
        {Name: "Class B", Values: []float64{65, 70, 78, 82, 88, 60, 75}},
    },
})
```

## Chart Options

| Option | Type | Description |
|--------|------|-------------|
| `Type` | `string` | Chart type (line, bar, scatter, pie, histogram, heatmap, box) |
| `Title` | `string` | Chart title |
| `Width` | `int` | Width in pixels (default: 600) |
| `Height` | `int` | Height in pixels (default: 400) |
| `XLabel` | `string` | X-axis label |
| `YLabel` | `string` | Y-axis label |
| `Series` | `[]Series` | Data series |
| `Bins` | `int` | Number of histogram bins |
| `Grid` | `[][]float64` | 2D data for heatmaps |
