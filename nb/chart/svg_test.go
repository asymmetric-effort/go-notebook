// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import (
	"strings"
	"testing"
)

func TestRenderBarSVG(t *testing.T) {
	c := Bar{
		Title:  "Revenue",
		Labels: []string{"Q1", "Q2", "Q3"},
		Series: []Series{{Name: "2024", Values: []float64{100, 200, 150}}},
	}
	svg := RenderBarSVG(c)
	if !strings.Contains(svg, "<svg") {
		t.Error("missing svg tag")
	}
	if !strings.Contains(svg, "Revenue") {
		t.Error("missing title")
	}
	if !strings.Contains(svg, "<rect") {
		t.Error("missing bars")
	}
}

func TestRenderBarSVGEmpty(t *testing.T) {
	c := Bar{Title: "Empty"}
	svg := RenderBarSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("should produce valid SVG even when empty")
	}
}

func TestRenderBarSVGEscaping(t *testing.T) {
	c := Bar{
		Labels: []string{"<script>alert('xss')</script>"},
		Series: []Series{{Name: "s", Values: []float64{1}}},
	}
	svg := RenderBarSVG(c)
	if strings.Contains(svg, "<script>") {
		t.Error("label should be escaped")
	}
}

func TestRenderLineSVG(t *testing.T) {
	c := Line{
		Title:  "Trend",
		X:      []float64{1, 2, 3, 4},
		Series: []Series{{Name: "data", Values: []float64{10, 20, 15, 25}}},
	}
	svg := RenderLineSVG(c)
	if !strings.Contains(svg, "<polyline") {
		t.Error("missing polyline")
	}
}

func TestRenderLineSVGEmpty(t *testing.T) {
	c := Line{Title: "Empty"}
	svg := RenderLineSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("should produce valid SVG")
	}
}

func TestRenderPieSVG(t *testing.T) {
	c := Pie{
		Title:  "Market Share",
		Labels: []string{"A", "B", "C"},
		Values: []float64{50, 30, 20},
	}
	svg := RenderPieSVG(c)
	if !strings.Contains(svg, "<path") {
		t.Error("missing pie slices")
	}
	if !strings.Contains(svg, "Market Share") {
		t.Error("missing title")
	}
}

func TestRenderPieSVGEmpty(t *testing.T) {
	c := Pie{Title: "Empty"}
	svg := RenderPieSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("should produce valid SVG")
	}
}

func TestRenderPieSVGZeroTotal(t *testing.T) {
	c := Pie{Labels: []string{"a"}, Values: []float64{0}}
	svg := RenderPieSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("should handle zero total")
	}
}

func TestRenderBarSVGZeroMax(t *testing.T) {
	c := Bar{Labels: []string{"a"}, Series: []Series{{Name: "s", Values: []float64{0}}}}
	svg := RenderBarSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("should handle zero max value")
	}
}

func TestRenderLineSVGSinglePoint(t *testing.T) {
	c := Line{X: []float64{1}, Series: []Series{{Name: "s", Values: []float64{5}}}}
	svg := RenderLineSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("should handle single point")
	}
}

func TestRenderSVGDispatcher(t *testing.T) {
	charts := []interface{}{
		Line{X: []float64{1}, Series: []Series{{Values: []float64{1}}}},
		Area{X: []float64{1}, Series: []Series{{Values: []float64{1}}}},
		Bar{Labels: []string{"a"}, Series: []Series{{Values: []float64{1}}}},
		Scatter{X: []float64{1}, Y: []float64{1}},
		Histogram{Values: []float64{1, 2, 3}},
		Pie{Labels: []string{"a"}, Values: []float64{1}},
		Heatmap{RowLabels: []string{"r"}, ColLabels: []string{"c"}, Values: [][]float64{{1}}},
	}
	for _, c := range charts {
		svg := RenderSVG(c)
		if !strings.Contains(svg, "<svg") {
			t.Errorf("RenderSVG(%T) missing <svg>", c)
		}
	}
}

func TestRenderSVGUnknownPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown type")
		}
	}()
	RenderSVG(42)
}
