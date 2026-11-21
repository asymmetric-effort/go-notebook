// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import (
	"strings"
	"testing"
)

func TestRenderLineSVGAllZeroY(t *testing.T) {
	// Triggers maxY == 0 branch.
	c := Line{X: []float64{1, 2, 3}, Series: []Series{{Name: "s", Values: []float64{0, 0, 0}}}}
	svg := RenderLineSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("should produce valid SVG")
	}
}

func TestRenderPieSVGLargeSlice(t *testing.T) {
	// A single slice > 180 degrees triggers largeArc = 1.
	c := Pie{Labels: []string{"big", "small"}, Values: []float64{90, 10}}
	svg := RenderPieSVG(c)
	if !strings.Contains(svg, "<path") {
		t.Error("missing pie path")
	}
}

func TestRenderAreaSVGAllZeroY(t *testing.T) {
	// Triggers maxY == 0 branch.
	c := Area{X: []float64{1, 2}, Series: []Series{{Name: "s", Values: []float64{0, 0}}}}
	svg := RenderAreaSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("should produce valid SVG")
	}
}

func TestRenderHistogramSVGAllSameValue(t *testing.T) {
	// All same values: maxVal == minVal triggers the guard.
	c := Histogram{Values: []float64{5, 5, 5, 5}}
	svg := RenderHistogramSVG(c)
	if !strings.Contains(svg, "<rect") {
		t.Error("should produce bars")
	}
}

func TestRenderHistogramSVGMaxCountZero(t *testing.T) {
	// Very small dataset with 1 value triggers numBins edge.
	c := Histogram{Values: []float64{1}}
	svg := RenderHistogramSVG(c)
	if !strings.Contains(svg, "</svg>") {
		t.Error("should produce valid SVG")
	}
}

func TestRenderScatterSVGEqualRange(t *testing.T) {
	// All same X and Y values: maxX==minX and maxY==minY.
	c := Scatter{X: []float64{5, 5}, Y: []float64{3, 3}}
	svg := RenderScatterSVG(c)
	if !strings.Contains(svg, "<circle") {
		t.Error("should produce circles")
	}
}
