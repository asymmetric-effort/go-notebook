// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import (
	"strings"
	"testing"
)

func TestRenderAreaSVG(t *testing.T) {
	c := Area{
		Title:  "Revenue",
		X:      []float64{1, 2, 3, 4},
		Series: []Series{{Name: "2024", Values: []float64{10, 20, 15, 25}}},
	}
	svg := RenderAreaSVG(c)
	if !strings.Contains(svg, "<polygon") {
		t.Error("missing polygon for area fill")
	}
	if !strings.Contains(svg, "Revenue") {
		t.Error("missing title")
	}
}

func TestRenderAreaSVGEmpty(t *testing.T) {
	svg := RenderAreaSVG(Area{Title: "Empty"})
	if !strings.Contains(svg, "</svg>") {
		t.Error("empty area should produce valid SVG")
	}
}

func TestRenderScatterSVG(t *testing.T) {
	c := Scatter{
		Title: "Correlation",
		X:     []float64{1, 2, 3, 4, 5},
		Y:     []float64{2, 4, 3, 5, 4},
	}
	svg := RenderScatterSVG(c)
	if !strings.Contains(svg, "<circle") {
		t.Error("missing circles for scatter points")
	}
}

func TestRenderScatterSVGWithGroups(t *testing.T) {
	c := Scatter{
		X:      []float64{1, 2, 3},
		Y:      []float64{1, 2, 3},
		Groups: []string{"a", "b", "a"},
	}
	svg := RenderScatterSVG(c)
	if strings.Count(svg, "<circle") != 3 {
		t.Errorf("expected 3 circles, got %d", strings.Count(svg, "<circle"))
	}
}

func TestRenderScatterSVGEmpty(t *testing.T) {
	svg := RenderScatterSVG(Scatter{})
	if !strings.Contains(svg, "</svg>") {
		t.Error("empty scatter should produce valid SVG")
	}
}

func TestRenderHistogramSVG(t *testing.T) {
	c := Histogram{
		Title:  "Distribution",
		Values: []float64{1, 2, 2, 3, 3, 3, 4, 4, 5},
	}
	svg := RenderHistogramSVG(c)
	if !strings.Contains(svg, "<rect") {
		t.Error("missing bars for histogram")
	}
	if !strings.Contains(svg, "Distribution") {
		t.Error("missing title")
	}
}

func TestRenderHistogramSVGWithBins(t *testing.T) {
	c := Histogram{
		Values: []float64{1, 2, 3, 4, 5},
		Bins:   3,
	}
	svg := RenderHistogramSVG(c)
	if strings.Count(svg, "<rect") != 3 {
		t.Errorf("expected 3 bins, got %d rects", strings.Count(svg, "<rect"))
	}
}

func TestRenderHistogramSVGEmpty(t *testing.T) {
	svg := RenderHistogramSVG(Histogram{})
	if !strings.Contains(svg, "</svg>") {
		t.Error("empty histogram should produce valid SVG")
	}
}

func TestRenderHistogramSVGSingleValue(t *testing.T) {
	c := Histogram{Values: []float64{5, 5, 5}}
	svg := RenderHistogramSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("single-value histogram should produce valid SVG")
	}
}

func TestRenderHeatmapSVG(t *testing.T) {
	c := Heatmap{
		Title:     "Correlation Matrix",
		RowLabels: []string{"A", "B"},
		ColLabels: []string{"X", "Y", "Z"},
		Values:    [][]float64{{1, 5, 3}, {4, 2, 6}},
	}
	svg := RenderHeatmapSVG(c)
	if !strings.Contains(svg, "<rect") {
		t.Error("missing cells for heatmap")
	}
	if strings.Count(svg, "<rect") != 6 {
		t.Errorf("expected 6 cells (2x3), got %d", strings.Count(svg, "<rect"))
	}
	if !strings.Contains(svg, "Correlation Matrix") {
		t.Error("missing title")
	}
}

func TestRenderHeatmapSVGEmpty(t *testing.T) {
	svg := RenderHeatmapSVG(Heatmap{})
	if !strings.Contains(svg, "</svg>") {
		t.Error("empty heatmap should produce valid SVG")
	}
}

func TestRenderHeatmapSVGEscaping(t *testing.T) {
	c := Heatmap{
		RowLabels: []string{"<script>"},
		ColLabels: []string{"normal"},
		Values:    [][]float64{{1}},
	}
	svg := RenderHeatmapSVG(c)
	if strings.Contains(svg, "<script>") {
		t.Error("label should be escaped")
	}
}

func TestRenderAreaSVGEscaping(t *testing.T) {
	c := Area{
		Title:  "<script>alert(1)</script>",
		X:      []float64{1},
		Series: []Series{{Name: "s", Values: []float64{1}}},
	}
	svg := RenderAreaSVG(c)
	if strings.Contains(svg, "<script>") {
		t.Error("title should be escaped")
	}
}

func TestRenderScatterSVGSinglePoint(t *testing.T) {
	c := Scatter{X: []float64{1}, Y: []float64{1}}
	svg := RenderScatterSVG(c)
	if !strings.Contains(svg, "<circle") {
		t.Error("single point should render")
	}
}
