// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.
//
// Typst WASM module for go-notebook. Provides compile-to-PDF and
// SVG-to-PNG conversion, accessed from Go via wazero.
//
// This module communicates through stdin/stdout: the Go host writes
// a JSON command to stdin and reads the result from stdout.

use std::io::{self, Read, Write};

/// Command protocol: JSON on stdin, result on stdout.
/// Commands:
///   {"cmd":"compile","input":"<typst source>","format":"pdf"}
///   {"cmd":"svg_to_png","input":"<svg content>","width":800,"height":600}

fn main() {
    let mut input = String::new();
    io::stdin().read_to_string(&mut input).unwrap_or_default();

    // For now, echo the input length as a simple proof of life.
    // The full Typst integration will parse the JSON command and
    // invoke the Typst compiler.
    let response = format!("{{\"ok\":true,\"input_len\":{}}}\n", input.len());
    io::stdout().write_all(response.as_bytes()).unwrap_or_default();
}
