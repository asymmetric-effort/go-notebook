// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import (
	"strings"
	"testing"
)

func TestSvgOpenTagContainsDataChart(t *testing.T) {
	series := []Series{{Name: "sales", Values: []float64{10, 20, 30}}}
	tag := svgOpenTag(600, 400, "bar", "Test", []string{"A", "B", "C"}, nil, series)

	if !strings.Contains(tag, "data-chart=") {
		t.Error("SVG tag should contain data-chart attribute")
	}
	// JSON is HTML-escaped in the attribute, so quotes become &quot;
	if !strings.Contains(tag, "bar") {
		t.Error("data-chart should contain chart type")
	}
	if !strings.Contains(tag, "sales") {
		t.Error("data-chart should contain series name")
	}
}

func TestSvgOpenTagEscapesJSON(t *testing.T) {
	series := []Series{{Name: `test"quote`, Values: []float64{1}}}
	tag := svgOpenTag(600, 400, "line", `Title"Test`, nil, nil, series)

	// The JSON is HTML-escaped in the attribute.
	if !strings.Contains(tag, "data-chart=") {
		t.Error("should have data-chart attribute")
	}
	// Should not contain raw quotes inside the attribute value.
	if strings.Contains(tag, `data-chart="{"`) {
		t.Error("JSON should be HTML-escaped")
	}
}

func TestSvgOpenTagNoSeries(t *testing.T) {
	tag := svgOpenTag(400, 400, "heatmap", "Heat", []string{"A"}, nil, nil)
	if !strings.Contains(tag, "data-chart=") {
		t.Error("should have data-chart even without series")
	}
}

func TestBarSVGContainsDataChart(t *testing.T) {
	c := Bar{
		Title:  "Sales",
		Labels: []string{"Q1", "Q2"},
		Series: []Series{{Name: "Revenue", Values: []float64{100, 200}}},
	}
	svg := RenderBarSVG(c)
	if !strings.Contains(svg, "data-chart=") {
		t.Error("bar SVG should embed chart data")
	}
	if !strings.Contains(svg, "<title>") {
		t.Error("bar SVG should have title elements for hover")
	}
}

func TestLineSVGContainsDataChart(t *testing.T) {
	c := Line{
		Title:  "Trend",
		X:      []float64{1, 2, 3},
		Series: []Series{{Name: "Growth", Values: []float64{10, 20, 30}}},
	}
	svg := RenderLineSVG(c)
	if !strings.Contains(svg, "data-chart=") {
		t.Error("line SVG should embed chart data")
	}
}

func TestPieSVGContainsDataChart(t *testing.T) {
	c := Pie{
		Title:  "Share",
		Labels: []string{"A", "B"},
		Values: []float64{60, 40},
	}
	svg := RenderPieSVG(c)
	if !strings.Contains(svg, "data-chart=") {
		t.Error("pie SVG should embed chart data")
	}
}

func TestChartScriptEmbedded(t *testing.T) {
	if ChartScript == "" {
		t.Fatal("ChartScript should be embedded from chart.js")
	}
	if !strings.Contains(ChartScript, "initCharts") {
		t.Error("ChartScript should contain initCharts function")
	}
	if !strings.Contains(ChartScript, "ChartInteraction") {
		t.Error("ChartScript should contain ChartInteraction class")
	}
}

func TestChartScriptNoEval(t *testing.T) {
	// The script must never use eval, Function(), or other code execution.
	if strings.Contains(ChartScript, "eval(") {
		t.Error("ChartScript must never use eval()")
	}
	if strings.Contains(ChartScript, "new Function(") {
		t.Error("ChartScript must never use new Function()")
	}
	if strings.Contains(ChartScript, "innerHTML") {
		t.Error("ChartScript should not use innerHTML (XSS risk)")
	}
}
