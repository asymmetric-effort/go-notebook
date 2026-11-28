# Charts

go-notebook includes a built-in chart rendering engine. Charts are created
using the `ctx.Chart()` method inside notebook cells and render as SVG in HTML
output or inline in PDF.

## Supported Chart Types

The chart package (`nb/chart`) provides seven typed structs:

| Struct | Description |
| ------ | ----------- |
| `chart.Line` | Line chart with X values and one or more series |
| `chart.Area` | Area chart, optionally stacked |
| `chart.Bar` | Bar chart with category labels and series |
| `chart.Scatter` | Scatter plot with X/Y data points |
| `chart.Histogram` | Frequency distribution from raw values |
| `chart.Pie` | Pie chart with labeled slices |
| `chart.Heatmap` | 2D grid with color intensity |

## Basic Usage

Pass a typed struct to `ctx.Chart()`. The chart is rendered as SVG and
displayed inline in the report output.

```go
ctx.Chart(chart.Line{
    Title:  "Monthly Revenue",
    XLabel: "Month",
    YLabel: "Revenue ($)",
    X:      []float64{1, 2, 3, 4, 5, 6},
    Series: []chart.Series{
        {Name: "2025", Values: []float64{10, 25, 30, 45, 50, 65}},
    },
})
```

## Line Chart

Line charts display data points connected by straight lines. Use them for
time series or continuous data.

```go
ctx.Chart(chart.Line{
    Title:  "Temperature Over Time",
    XLabel: "Hour",
    YLabel: "Celsius",
    X:      []float64{0, 4, 8, 12, 16, 20},
    Series: []chart.Series{
        {Name: "Sensor A", Values: []float64{20, 22, 19, 25, 28, 24}},
        {Name: "Sensor B", Values: []float64{18, 20, 21, 23, 26, 22}},
    },
})
```

### chart.Line fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Title` | `string` | Chart title |
| `XLabel` | `string` | X-axis label |
| `YLabel` | `string` | Y-axis label |
| `X` | `[]float64` | X-axis values |
| `Series` | `[]chart.Series` | Named data series |

## Area Chart

Area charts are like line charts with the area under the curve filled.
Set `Stacked: true` for stacked areas.

```go
ctx.Chart(chart.Area{
    Title:   "Traffic by Source",
    XLabel:  "Day",
    YLabel:  "Visitors",
    X:       []float64{1, 2, 3, 4, 5},
    Stacked: true,
    Series: []chart.Series{
        {Name: "Organic", Values: []float64{100, 120, 140, 130, 160}},
        {Name: "Direct", Values: []float64{50, 60, 55, 70, 80}},
    },
})
```

### chart.Area fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Title` | `string` | Chart title |
| `XLabel` | `string` | X-axis label |
| `YLabel` | `string` | Y-axis label |
| `X` | `[]float64` | X-axis values |
| `Series` | `[]chart.Series` | Named data series |
| `Stacked` | `bool` | Stack areas on top of each other |

## Bar Chart

Bar charts compare discrete categories. Set `Stacked: true` for stacked bars.

```go
ctx.Chart(chart.Bar{
    Title:  "Sales by Region",
    Labels: []string{"North", "South", "East", "West"},
    Series: []chart.Series{
        {Name: "Q1", Values: []float64{150, 230, 180, 310}},
        {Name: "Q2", Values: []float64{200, 210, 190, 280}},
    },
})
```

### chart.Bar fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Title` | `string` | Chart title |
| `Labels` | `[]string` | Category labels |
| `Series` | `[]chart.Series` | Named data series |
| `Stacked` | `bool` | Stack bars on top of each other |

## Scatter Plot

Scatter plots show the relationship between two variables.

```go
ctx.Chart(chart.Scatter{
    Title:  "Height vs Weight",
    XLabel: "Height (cm)",
    YLabel: "Weight (kg)",
    X:      []float64{160, 170, 175, 180},
    Y:      []float64{55, 70, 75, 85},
    Groups: []string{"M", "F", "M", "M"},
})
```

### chart.Scatter fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Title` | `string` | Chart title |
| `XLabel` | `string` | X-axis label |
| `YLabel` | `string` | Y-axis label |
| `X` | `[]float64` | X-axis values |
| `Y` | `[]float64` | Y-axis values |
| `Groups` | `[]string` | Optional group labels for coloring |

## Histogram

Histograms compute bins from raw values to show frequency distributions.

```go
ctx.Chart(chart.Histogram{
    Title:  "Response Times",
    XLabel: "Latency (ms)",
    YLabel: "Frequency",
    Values: []float64{12, 15, 14, 18, 22, 19, 25, 30, 28, 35},
    Bins:   5,
})
```

### chart.Histogram fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Title` | `string` | Chart title |
| `XLabel` | `string` | X-axis label |
| `YLabel` | `string` | Y-axis label |
| `Values` | `[]float64` | Raw data values |
| `Bins` | `int` | Number of bins (0 means auto) |

## Pie Chart

Pie charts show proportional composition.

```go
ctx.Chart(chart.Pie{
    Title:  "Market Share",
    Labels: []string{"Product A", "Product B", "Product C", "Other"},
    Values: []float64{45, 25, 20, 10},
})
```

### chart.Pie fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Title` | `string` | Chart title |
| `Labels` | `[]string` | Slice labels |
| `Values` | `[]float64` | Slice values |

## Heatmap

Heatmaps display 2D data using color intensity.

```go
ctx.Chart(chart.Heatmap{
    Title:     "Correlation Matrix",
    RowLabels: []string{"A", "B", "C"},
    ColLabels: []string{"A", "B", "C"},
    Values: [][]float64{
        {1.0, 0.8, 0.3},
        {0.8, 1.0, 0.5},
        {0.3, 0.5, 1.0},
    },
})
```

### chart.Heatmap fields

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Title` | `string` | Chart title |
| `RowLabels` | `[]string` | Row labels |
| `ColLabels` | `[]string` | Column labels |
| `Values` | `[][]float64` | 2D data matrix |

## chart.Series

The `Series` struct is used by Line, Area, and Bar charts.

| Field | Type | Description |
| ----- | ---- | ----------- |
| `Name` | `string` | Series name (shown in legend) |
| `Values` | `[]float64` | Data values |
