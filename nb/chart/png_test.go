// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/engine"
)

func TestRenderPNGBar(t *testing.T) {
	ctx := context.Background()
	eng, err := engine.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close(ctx)

	c := Bar{
		Title:  "Test",
		Labels: []string{"A", "B"},
		Series: []Series{{Name: "s1", Values: []float64{10, 20}}},
	}

	b64, err := RenderPNG(ctx, eng, c, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	if b64 == "" {
		t.Error("PNG base64 should not be empty")
	}

	// Decode and check PNG magic.
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 8 || string(data[1:4]) != "PNG" {
		t.Errorf("not a PNG: bytes 1-3 = %q", data[1:4])
	}
	t.Logf("Chart PNG: %d bytes", len(data))
}

func TestRenderPNGLine(t *testing.T) {
	ctx := context.Background()
	eng, err := engine.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close(ctx)

	c := Line{
		Title:  "Revenue",
		X:      []float64{1, 2, 3},
		Series: []Series{{Name: "2024", Values: []float64{100, 200, 150}}},
	}

	b64, err := RenderPNG(ctx, eng, c, 400, 300)
	if err != nil {
		t.Fatal(err)
	}
	if b64 == "" {
		t.Error("PNG should not be empty")
	}
}
