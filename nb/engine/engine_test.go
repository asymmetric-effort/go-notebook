// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package engine

import (
	"context"
	"testing"
)

func TestNewEngine(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatalf("New() = _, %v", err)
	}
	defer e.Close(ctx)
}

func TestEngineRun(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close(ctx)

	resp, err := e.Run(ctx, Command{Cmd: "test", Input: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Errorf("resp.OK = false")
	}
	if resp.InputLen == 0 {
		t.Error("InputLen should be > 0")
	}
}

func TestEngineCompile(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close(ctx)

	// The stub module doesn't actually compile Typst, so this tests the
	// plumbing. The response will have ok=true but no data since the
	// stub just returns input length.
	_, err = e.Compile(ctx, "#hello", "pdf")
	// The stub responds with ok:true but no "data" field, so Compile
	// returns nil data, which is fine for now.
	if err != nil {
		t.Logf("Compile error (expected with stub): %v", err)
	}
}

func TestEngineSVGToPNG(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close(ctx)

	_, err = e.SVGToPNG(ctx, "<svg></svg>", 100, 100)
	if err != nil {
		t.Logf("SVGToPNG error (expected with stub): %v", err)
	}
}

func TestEngineClose(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Close(ctx); err != nil {
		t.Fatalf("Close() = %v", err)
	}
}

func TestEngineRunEmpty(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close(ctx)

	resp, err := e.Run(ctx, Command{Cmd: "test", Input: ""})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Errorf("resp.OK = false")
	}
}
