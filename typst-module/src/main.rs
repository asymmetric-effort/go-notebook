// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.
//
// Typst WASM module for go-notebook. Provides compile-to-PDF,
// accessed from Go via wazero.
//
// This module communicates through stdin/stdout: the Go host writes
// a JSON command to stdin and reads the result from stdout.

use base64::Engine as _;
use serde::{Deserialize, Serialize};
use std::io::{self, Read, Write};
use std::path::PathBuf;
use typst::foundations::{Bytes, Datetime};
use typst::layout::PagedDocument;
use typst::syntax::{FileId, Source};
use typst::text::{Font, FontBook};
use typst::utils::LazyHash;
use typst::Library;
use typst_library::World;

#[derive(Deserialize)]
struct Command {
    cmd: String,
    input: String,
    #[serde(default)]
    format: String,
    #[serde(default)]
    width: Option<u32>,
    #[serde(default)]
    height: Option<u32>,
}

#[derive(Serialize)]
struct Response {
    ok: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<String>, // base64-encoded
    #[serde(skip_serializing_if = "Option::is_none")]
    input_len: Option<usize>,
}

/// Minimal World implementation for WASM — no fonts, no files.
struct WasmWorld {
    library: LazyHash<Library>,
    book: LazyHash<FontBook>,
    source: Source,
    fonts: Vec<Font>,
}

impl WasmWorld {
    fn new(source_text: &str) -> Self {
        let main_id = FileId::new(None, typst::syntax::VirtualPath::new("main.typ"));
        Self {
            library: LazyHash::new(Library::default()),
            book: LazyHash::new(FontBook::new()),
            source: Source::new(main_id, source_text.into()),
            fonts: Vec::new(),
        }
    }
}

impl World for WasmWorld {
    fn library(&self) -> &LazyHash<Library> {
        &self.library
    }

    fn book(&self) -> &LazyHash<FontBook> {
        &self.book
    }

    fn main(&self) -> FileId {
        self.source.id()
    }

    fn source(&self, id: FileId) -> typst::diag::FileResult<Source> {
        if id == self.source.id() {
            Ok(self.source.clone())
        } else {
            Err(typst::diag::FileError::NotFound(PathBuf::from(id.vpath().as_rootless_path())))
        }
    }

    fn file(&self, id: FileId) -> typst::diag::FileResult<Bytes> {
        Err(typst::diag::FileError::NotFound(PathBuf::from(id.vpath().as_rootless_path())))
    }

    fn font(&self, index: usize) -> Option<Font> {
        self.fonts.get(index).cloned()
    }

    fn today(&self, _offset: Option<i64>) -> Option<Datetime> {
        Some(Datetime::from_ymd(2026, 1, 1).unwrap())
    }
}

fn compile_to_pdf(source: &str) -> Result<Vec<u8>, String> {
    let world = WasmWorld::new(source);
    let result = typst::compile::<PagedDocument>(&world);
    match result.output {
        Ok(doc) => {
            let options = typst_pdf::PdfOptions::default();
            match typst_pdf::pdf(&doc, &options) {
                Ok(pdf_bytes) => Ok(pdf_bytes),
                Err(diagnostics) => {
                    let msgs: Vec<String> = diagnostics
                        .iter()
                        .map(|d| d.message.to_string())
                        .collect();
                    Err(format!("pdf generation: {}", msgs.join("; ")))
                }
            }
        }
        Err(diagnostics) => {
            let msgs: Vec<String> = diagnostics
                .iter()
                .map(|d| d.message.to_string())
                .collect();
            Err(msgs.join("; "))
        }
    }
}

fn svg_to_png(svg_data: &str, width: u32, height: u32) -> Result<Vec<u8>, String> {
    let opts = usvg::Options::default();
    let tree = usvg::Tree::from_str(svg_data, &opts)
        .map_err(|e| format!("parse SVG: {}", e))?;

    let target_w = if width > 0 { width } else { 800 };
    let target_h = if height > 0 { height } else { 600 };

    let mut pixmap = tiny_skia::Pixmap::new(target_w, target_h)
        .ok_or_else(|| "failed to create pixmap".to_string())?;

    let tree_size = tree.size();
    let sx = target_w as f32 / tree_size.width();
    let sy = target_h as f32 / tree_size.height();
    let scale = sx.min(sy);

    resvg::render(
        &tree,
        tiny_skia::Transform::from_scale(scale, scale),
        &mut pixmap.as_mut(),
    );

    // Encode to PNG.
    let mut png_buf = Vec::new();
    {
        let mut encoder = png::Encoder::new(&mut png_buf, target_w, target_h);
        encoder.set_color(png::ColorType::Rgba);
        encoder.set_depth(png::BitDepth::Eight);
        let mut writer = encoder.write_header()
            .map_err(|e| format!("PNG header: {}", e))?;
        writer.write_image_data(pixmap.data())
            .map_err(|e| format!("PNG write: {}", e))?;
    }

    Ok(png_buf)
}

fn main() {
    let mut input = String::new();
    io::stdin().read_to_string(&mut input).unwrap_or_default();

    let response = match serde_json::from_str::<Command>(&input) {
        Ok(cmd) => match cmd.cmd.as_str() {
            "compile" => match cmd.format.as_str() {
                "pdf" | "" => match compile_to_pdf(&cmd.input) {
                    Ok(pdf_bytes) => {
                        let encoded = base64::engine::general_purpose::STANDARD.encode(&pdf_bytes);
                        Response {
                            ok: true,
                            error: None,
                            data: Some(encoded),
                            input_len: None,
                        }
                    }
                    Err(e) => Response {
                        ok: false,
                        error: Some(e),
                        data: None,
                        input_len: None,
                    },
                },
                other => Response {
                    ok: false,
                    error: Some(format!("unsupported format: {}", other)),
                    data: None,
                    input_len: None,
                },
            },
            "svg_to_png" => {
                let w = cmd.width.unwrap_or(800) as u32;
                let h = cmd.height.unwrap_or(600) as u32;
                match svg_to_png(&cmd.input, w, h) {
                    Ok(png_bytes) => {
                        let encoded = base64::engine::general_purpose::STANDARD.encode(&png_bytes);
                        Response {
                            ok: true,
                            error: None,
                            data: Some(encoded),
                            input_len: None,
                        }
                    }
                    Err(e) => Response {
                        ok: false,
                        error: Some(e),
                        data: None,
                        input_len: None,
                    },
                }
            },
            "ping" => Response {
                ok: true,
                error: None,
                data: None,
                input_len: Some(cmd.input.len()),
            },
            other => Response {
                ok: false,
                error: Some(format!("unknown command: {}", other)),
                data: None,
                input_len: None,
            },
        },
        Err(e) => Response {
            ok: false,
            error: Some(format!("parse error: {}", e)),
            data: None,
            input_len: None,
        },
    };

    let json = serde_json::to_string(&response).unwrap_or_else(|_| r#"{"ok":false}"#.into());
    io::stdout().write_all(json.as_bytes()).unwrap_or_default();
    io::stdout().write_all(b"\n").unwrap_or_default();
}
