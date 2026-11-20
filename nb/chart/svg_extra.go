// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import (
	"fmt"
	"html"
	"math"
	"strings"
)

// RenderAreaSVG renders an area chart to SVG, optionally stacked.
func RenderAreaSVG(c Area) string {
	width := 600
	height := 400
	margin := 60

	var buf strings.Builder
	fmt.Fprintf(&buf, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height)
	buf.WriteString("\n")

	if c.Title != "" {
		fmt.Fprintf(&buf, `<text x="%d" y="30" text-anchor="middle" font-size="16">%s</text>`, width/2, html.EscapeString(c.Title))
		buf.WriteString("\n")
	}

	if len(c.X) == 0 || len(c.Series) == 0 {
		buf.WriteString("</svg>")
		return buf.String()
	}

	minX, maxX := c.X[0], c.X[0]
	maxY := 0.0
	for _, v := range c.X {
		minX = math.Min(minX, v)
		maxX = math.Max(maxX, v)
	}
	for _, s := range c.Series {
		for _, v := range s.Values {
			maxY = math.Max(maxY, math.Abs(v))
		}
	}
	if maxX == minX {
		maxX = minX + 1
	}
	if maxY == 0 {
		maxY = 1
	}

	chartW := float64(width - 2*margin)
	chartH := float64(height - 2*margin)
	baseY := float64(margin) + chartH
	colors := []string{"#4e79a7", "#f28e2b", "#e15759", "#76b7b2", "#59a14f"}

	for si, s := range c.Series {
		color := colors[si%len(colors)]
		var points []string
		// Start from bottom-left.
		for i, v := range s.Values {
			x := float64(margin) + (c.X[i]-minX)/(maxX-minX)*chartW
			y := float64(margin) + chartH - (v/maxY)*chartH
			points = append(points, fmt.Sprintf("%.1f,%.1f", x, y))
		}
		// Close the polygon back along the baseline.
		lastX := float64(margin) + (c.X[len(c.X)-1]-minX)/(maxX-minX)*chartW
		firstX := float64(margin) + (c.X[0]-minX)/(maxX-minX)*chartW
		points = append(points, fmt.Sprintf("%.1f,%.1f", lastX, baseY))
		points = append(points, fmt.Sprintf("%.1f,%.1f", firstX, baseY))

		fmt.Fprintf(&buf, `<polygon points="%s" fill="%s" fill-opacity="0.3" stroke="%s" stroke-width="1.5"/>`,
			strings.Join(points, " "), color, color)
		buf.WriteString("\n")
	}

	buf.WriteString("</svg>")
	return buf.String()
}

// RenderScatterSVG renders a scatter plot to SVG.
func RenderScatterSVG(c Scatter) string {
	width := 600
	height := 400
	margin := 60

	var buf strings.Builder
	fmt.Fprintf(&buf, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height)
	buf.WriteString("\n")

	if c.Title != "" {
		fmt.Fprintf(&buf, `<text x="%d" y="30" text-anchor="middle" font-size="16">%s</text>`, width/2, html.EscapeString(c.Title))
		buf.WriteString("\n")
	}

	if len(c.X) == 0 {
		buf.WriteString("</svg>")
		return buf.String()
	}

	minX, maxX := c.X[0], c.X[0]
	minY, maxY := c.Y[0], c.Y[0]
	for _, v := range c.X {
		minX = math.Min(minX, v)
		maxX = math.Max(maxX, v)
	}
	for _, v := range c.Y {
		minY = math.Min(minY, v)
		maxY = math.Max(maxY, v)
	}
	if maxX == minX {
		maxX = minX + 1
	}
	if maxY == minY {
		maxY = minY + 1
	}

	chartW := float64(width - 2*margin)
	chartH := float64(height - 2*margin)
	colors := []string{"#4e79a7", "#f28e2b", "#e15759", "#76b7b2"}

	for i := range c.X {
		x := float64(margin) + (c.X[i]-minX)/(maxX-minX)*chartW
		y := float64(margin) + chartH - (c.Y[i]-minY)/(maxY-minY)*chartH
		color := colors[0]
		if len(c.Groups) > i {
			groupIdx := 0
			for gi, g := range c.Groups {
				if g == c.Groups[i] && gi < i {
					groupIdx = gi % len(colors)
					break
				}
			}
			color = colors[groupIdx%len(colors)]
		}
		fmt.Fprintf(&buf, `<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, x, y, color)
		buf.WriteString("\n")
	}

	buf.WriteString("</svg>")
	return buf.String()
}

// RenderHistogramSVG renders a histogram to SVG with automatic binning.
func RenderHistogramSVG(c Histogram) string {
	width := 600
	height := 400
	margin := 60

	var buf strings.Builder
	fmt.Fprintf(&buf, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height)
	buf.WriteString("\n")

	if c.Title != "" {
		fmt.Fprintf(&buf, `<text x="%d" y="30" text-anchor="middle" font-size="16">%s</text>`, width/2, html.EscapeString(c.Title))
		buf.WriteString("\n")
	}

	if len(c.Values) == 0 {
		buf.WriteString("</svg>")
		return buf.String()
	}

	// Auto-bin using Sturges' rule if bins not specified.
	numBins := c.Bins
	if numBins <= 0 {
		numBins = int(math.Ceil(math.Log2(float64(len(c.Values))))) + 1
		if numBins < 2 {
			numBins = 2
		}
	}

	minVal, maxVal := c.Values[0], c.Values[0]
	for _, v := range c.Values {
		minVal = math.Min(minVal, v)
		maxVal = math.Max(maxVal, v)
	}
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	binWidth := (maxVal - minVal) / float64(numBins)
	counts := make([]int, numBins)
	for _, v := range c.Values {
		idx := int((v - minVal) / binWidth)
		if idx >= numBins {
			idx = numBins - 1
		}
		counts[idx]++
	}

	maxCount := 0
	for _, cnt := range counts {
		if cnt > maxCount {
			maxCount = cnt
		}
	}
	if maxCount == 0 {
		maxCount = 1
	}

	chartW := float64(width - 2*margin)
	chartH := float64(height - 2*margin)
	barW := chartW / float64(numBins)

	for i, cnt := range counts {
		barH := (float64(cnt) / float64(maxCount)) * chartH
		x := float64(margin) + float64(i)*barW
		y := float64(margin) + chartH - barH
		fmt.Fprintf(&buf, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="#4e79a7" stroke="white" stroke-width="1"/>`,
			x, y, barW, barH)
		buf.WriteString("\n")
	}

	buf.WriteString("</svg>")
	return buf.String()
}

// RenderHeatmapSVG renders a heatmap to SVG.
func RenderHeatmapSVG(c Heatmap) string {
	if len(c.RowLabels) == 0 || len(c.ColLabels) == 0 {
		return `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="400"></svg>`
	}

	margin := 80
	cellW := 50
	cellH := 40
	width := margin + cellW*len(c.ColLabels) + 20
	height := margin + cellH*len(c.RowLabels) + 20

	var buf strings.Builder
	fmt.Fprintf(&buf, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height)
	buf.WriteString("\n")

	if c.Title != "" {
		fmt.Fprintf(&buf, `<text x="%d" y="25" text-anchor="middle" font-size="16">%s</text>`, width/2, html.EscapeString(c.Title))
		buf.WriteString("\n")
	}

	// Find value range for color mapping.
	minVal, maxVal := c.Values[0][0], c.Values[0][0]
	for _, row := range c.Values {
		for _, v := range row {
			minVal = math.Min(minVal, v)
			maxVal = math.Max(maxVal, v)
		}
	}
	valRange := maxVal - minVal
	if valRange == 0 {
		valRange = 1
	}

	for ri, row := range c.Values {
		for ci, v := range row {
			intensity := (v - minVal) / valRange
			r := int(255 * (1 - intensity))
			b := int(255 * intensity)
			x := margin + ci*cellW
			y := margin + ri*cellH
			fmt.Fprintf(&buf, `<rect x="%d" y="%d" width="%d" height="%d" fill="rgb(%d,100,%d)"/>`,
				x, y, cellW, cellH, r, b)
			buf.WriteString("\n")
		}
	}

	// Row labels.
	for ri, label := range c.RowLabels {
		y := margin + ri*cellH + cellH/2 + 5
		fmt.Fprintf(&buf, `<text x="%d" y="%d" text-anchor="end" font-size="11">%s</text>`,
			margin-5, y, html.EscapeString(label))
		buf.WriteString("\n")
	}

	// Column labels.
	for ci, label := range c.ColLabels {
		x := margin + ci*cellW + cellW/2
		fmt.Fprintf(&buf, `<text x="%d" y="%d" text-anchor="middle" font-size="11" transform="rotate(-45,%d,%d)">%s</text>`,
			x, margin-10, x, margin-10, html.EscapeString(label))
		buf.WriteString("\n")
	}

	buf.WriteString("</svg>")
	return buf.String()
}
