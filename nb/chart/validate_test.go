// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import "testing"

func TestValidateLinePass(t *testing.T) {
	c := Line{X: []float64{1, 2, 3}, Series: []Series{{Name: "a", Values: []float64{1, 2, 3}}}}
	if err := ValidateLine(c); err != nil {
		t.Fatal(err)
	}
}

func TestValidateLineMismatch(t *testing.T) {
	c := Line{X: []float64{1, 2}, Series: []Series{{Name: "a", Values: []float64{1, 2, 3}}}}
	if err := ValidateLine(c); err == nil {
		t.Error("expected error for mismatched series length")
	}
}

func TestValidateAreaPass(t *testing.T) {
	c := Area{X: []float64{1, 2}, Series: []Series{{Name: "a", Values: []float64{1, 2}}}}
	if err := ValidateArea(c); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAreaMismatch(t *testing.T) {
	c := Area{X: []float64{1}, Series: []Series{{Name: "a", Values: []float64{1, 2}}}}
	if err := ValidateArea(c); err == nil {
		t.Error("expected error")
	}
}

func TestValidateBarPass(t *testing.T) {
	c := Bar{Labels: []string{"a", "b"}, Series: []Series{{Name: "s", Values: []float64{1, 2}}}}
	if err := ValidateBar(c); err != nil {
		t.Fatal(err)
	}
}

func TestValidateBarMismatch(t *testing.T) {
	c := Bar{Labels: []string{"a"}, Series: []Series{{Name: "s", Values: []float64{1, 2}}}}
	if err := ValidateBar(c); err == nil {
		t.Error("expected error")
	}
}

func TestValidateScatterPass(t *testing.T) {
	c := Scatter{X: []float64{1, 2}, Y: []float64{3, 4}}
	if err := ValidateScatter(c); err != nil {
		t.Fatal(err)
	}
}

func TestValidateScatterMismatch(t *testing.T) {
	c := Scatter{X: []float64{1, 2}, Y: []float64{3}}
	if err := ValidateScatter(c); err == nil {
		t.Error("expected error")
	}
}

func TestValidateScatterGroupsMismatch(t *testing.T) {
	c := Scatter{X: []float64{1, 2}, Y: []float64{3, 4}, Groups: []string{"a"}}
	if err := ValidateScatter(c); err == nil {
		t.Error("expected error for groups mismatch")
	}
}

func TestValidateHistogramPass(t *testing.T) {
	c := Histogram{Values: []float64{1, 2, 3}}
	if err := ValidateHistogram(c); err != nil {
		t.Fatal(err)
	}
}

func TestValidateHistogramEmpty(t *testing.T) {
	c := Histogram{}
	if err := ValidateHistogram(c); err == nil {
		t.Error("expected error for empty histogram")
	}
}

func TestValidatePiePass(t *testing.T) {
	c := Pie{Labels: []string{"a", "b"}, Values: []float64{30, 70}}
	if err := ValidatePie(c); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePieMismatch(t *testing.T) {
	c := Pie{Labels: []string{"a"}, Values: []float64{30, 70}}
	if err := ValidatePie(c); err == nil {
		t.Error("expected error")
	}
}

func TestValidatePieNegative(t *testing.T) {
	c := Pie{Labels: []string{"a", "b"}, Values: []float64{30, -10}}
	if err := ValidatePie(c); err == nil {
		t.Error("expected error for negative value")
	}
}

func TestValidateHeatmapPass(t *testing.T) {
	c := Heatmap{
		RowLabels: []string{"r1", "r2"},
		ColLabels: []string{"c1", "c2"},
		Values:    [][]float64{{1, 2}, {3, 4}},
	}
	if err := ValidateHeatmap(c); err != nil {
		t.Fatal(err)
	}
}

func TestValidateHeatmapRowMismatch(t *testing.T) {
	c := Heatmap{
		RowLabels: []string{"r1"},
		ColLabels: []string{"c1"},
		Values:    [][]float64{{1}, {2}},
	}
	if err := ValidateHeatmap(c); err == nil {
		t.Error("expected error for row count mismatch")
	}
}

func TestValidateHeatmapColMismatch(t *testing.T) {
	c := Heatmap{
		RowLabels: []string{"r1"},
		ColLabels: []string{"c1", "c2"},
		Values:    [][]float64{{1}},
	}
	if err := ValidateHeatmap(c); err == nil {
		t.Error("expected error for column count mismatch")
	}
}
