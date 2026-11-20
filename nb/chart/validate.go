// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import "fmt"

// Validate checks chart data for consistency. Returns a human-readable
// error if the data is malformed.
func ValidateLine(c Line) error {
	for _, s := range c.Series {
		if len(s.Values) != len(c.X) {
			return fmt.Errorf("series %q has %d values but X has %d", s.Name, len(s.Values), len(c.X))
		}
	}
	return nil
}

// ValidateArea validates an area chart.
func ValidateArea(c Area) error {
	for _, s := range c.Series {
		if len(s.Values) != len(c.X) {
			return fmt.Errorf("series %q has %d values but X has %d", s.Name, len(s.Values), len(c.X))
		}
	}
	return nil
}

// ValidateBar validates a bar chart.
func ValidateBar(c Bar) error {
	for _, s := range c.Series {
		if len(s.Values) != len(c.Labels) {
			return fmt.Errorf("series %q has %d values but %d labels", s.Name, len(s.Values), len(c.Labels))
		}
	}
	return nil
}

// ValidateScatter validates a scatter chart.
func ValidateScatter(c Scatter) error {
	if len(c.X) != len(c.Y) {
		return fmt.Errorf("X has %d values but Y has %d", len(c.X), len(c.Y))
	}
	if len(c.Groups) > 0 && len(c.Groups) != len(c.X) {
		return fmt.Errorf("groups has %d entries but X has %d", len(c.Groups), len(c.X))
	}
	return nil
}

// ValidateHistogram validates a histogram.
func ValidateHistogram(c Histogram) error {
	if len(c.Values) == 0 {
		return fmt.Errorf("histogram has no values")
	}
	return nil
}

// ValidatePie validates a pie chart.
func ValidatePie(c Pie) error {
	if len(c.Labels) != len(c.Values) {
		return fmt.Errorf("pie has %d labels but %d values", len(c.Labels), len(c.Values))
	}
	for i, v := range c.Values {
		if v < 0 {
			return fmt.Errorf("pie value %d (%q) is negative: %f", i, c.Labels[i], v)
		}
	}
	return nil
}

// ValidateHeatmap validates a heatmap.
func ValidateHeatmap(c Heatmap) error {
	if len(c.Values) != len(c.RowLabels) {
		return fmt.Errorf("heatmap has %d rows but %d row labels", len(c.Values), len(c.RowLabels))
	}
	for i, row := range c.Values {
		if len(row) != len(c.ColLabels) {
			return fmt.Errorf("heatmap row %d has %d values but %d column labels", i, len(row), len(c.ColLabels))
		}
	}
	return nil
}
