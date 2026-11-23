// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package chart

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/asymmetric-effort/go-notebook/nb/engine"
)

// RenderPNG renders any chart type to a base64-encoded PNG image using the
// Typst WASM engine for SVG-to-PNG conversion. Returns the PNG as base64.
func RenderPNG(ctx context.Context, eng *engine.Engine, c interface{}, width, height int) (string, error) {
	svg := RenderSVG(c)
	png, err := eng.SVGToPNG(ctx, svg, width, height)
	if err != nil {
		return "", fmt.Errorf("chart to PNG: %w", err)
	}
	return base64.StdEncoding.EncodeToString(png), nil
}
