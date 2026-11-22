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

func TestEnginePing(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close(ctx)

	resp, err := e.Run(ctx, Command{Cmd: "ping", Input: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Errorf("resp.OK = false, error: %s", resp.Error)
	}
	if resp.InputLen == 0 {
		t.Error("InputLen should be > 0")
	}
}

func TestEngineCompilePDF(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close(ctx)

	// Simple Typst source that should compile to PDF.
	source := "#set page(width: 100pt, height: 100pt)\nHello, world!"
	pdf, err := e.Compile(ctx, source, "pdf")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if len(pdf) == 0 {
		t.Error("PDF output should not be empty")
	}
	// Check PDF magic bytes.
	if len(pdf) > 4 && string(pdf[:5]) != "%PDF-" {
		t.Errorf("PDF should start with %%PDF-, got %q", pdf[:5])
	}
	t.Logf("PDF size: %d bytes", len(pdf))
}

func TestEngineCompileError(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close(ctx)

	// Invalid Typst source.
	_, err = e.Compile(ctx, "#nonexistent_function()", "pdf")
	if err == nil {
		t.Error("expected compile error for invalid source")
	}
	t.Logf("Expected error: %v", err)
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

func TestEngineUnknownCommand(t *testing.T) {
	ctx := context.Background()
	e, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close(ctx)

	resp, err := e.Run(ctx, Command{Cmd: "bogus", Input: ""})
	if err != nil {
		t.Fatal(err)
	}
	if resp.OK {
		t.Error("unknown command should not be OK")
	}
}
