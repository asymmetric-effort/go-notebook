// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package chart provides the chart types and renderer for go-notebook.
// Authors pass data to ctx.Chart and the tooling renders it.
package chart

import "fmt"

// Series holds a named data series for line, area and bar charts.
type Series struct {
	Name   string
	Values []float64
}

// Line is a line chart with X values and one or more series.
type Line struct {
	Title  string
	XLabel string
	YLabel string
	X      []float64
	Series []Series
}

// Area is an area chart, optionally stacked.
type Area struct {
	Title   string
	XLabel  string
	YLabel  string
	X       []float64
	Series  []Series
	Stacked bool
}

// Bar is a bar chart with category labels and series.
type Bar struct {
	Title   string
	Labels  []string
	Series  []Series
	Stacked bool
}

// Scatter is a scatter plot with X/Y values and optional groups.
type Scatter struct {
	Title  string
	XLabel string
	YLabel string
	X      []float64
	Y      []float64
	Groups []string
}

// Histogram computes bins from raw values.
type Histogram struct {
	Title  string
	XLabel string
	YLabel string
	Values []float64
	Bins   int // 0 means auto.
}

// Pie is a pie chart with category labels and values.
type Pie struct {
	Title  string
	Labels []string
	Values []float64
}

// Heatmap is a heatmap with row/column labels and a value matrix.
type Heatmap struct {
	Title     string
	RowLabels []string
	ColLabels []string
	Values    [][]float64
}

// RenderSVG renders any chart type to an SVG string. Panics on unknown types.
func RenderSVG(c interface{}) string {
	switch v := c.(type) {
	case Line:
		return RenderLineSVG(v)
	case Area:
		return RenderAreaSVG(v)
	case Bar:
		return RenderBarSVG(v)
	case Scatter:
		return RenderScatterSVG(v)
	case Histogram:
		return RenderHistogramSVG(v)
	case Pie:
		return RenderPieSVG(v)
	case Heatmap:
		return RenderHeatmapSVG(v)
	default:
		panic(fmt.Sprintf("unknown chart type: %T", c))
	}
}
