// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// chartDataJSON is the JSON structure embedded in the SVG data-chart attribute.
type chartDataJSON struct {
	Type   string            `json:"type"`
	Title  string            `json:"title"`
	Labels []string          `json:"labels,omitempty"`
	X      []float64         `json:"x,omitempty"`
	Series []chartSeriesJSON `json:"series,omitempty"`
}

type chartSeriesJSON struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
}

// svgOpenTag returns an SVG open tag with the chart data embedded as a
// data-chart attribute. The chart script reads this for interactivity.
func svgOpenTag(width, height int, chartType string, title string, labels []string, x []float64, series []Series) string {
	data := chartDataJSON{
		Type:   chartType,
		Title:  title,
		Labels: labels,
		X:      x,
	}
	for _, s := range series {
		data.Series = append(data.Series, chartSeriesJSON{
			Name:   s.Name,
			Values: s.Values,
		})
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		// Fallback: no data attribute.
		return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height)
	}
	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" data-chart="%s">`,
		width, height, width, height, encoded)
}

//go:embed static/chart.js
var ChartScript string
