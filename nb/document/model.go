// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package document provides the intermediate document model used by all
// report output backends (Markdown, HTML, PDF).
package document

// Document holds the complete report content in a backend-agnostic form.
type Document struct {
	Elements []Element
}

// ElementKind identifies the type of a document element.
type ElementKind int

const (
	KindProse   ElementKind = iota // Markdown prose text.
	KindDisplay                    // Cell display output (text/markdown, images).
	KindError                      // Error block at a failed cell's position.
	KindImage                      // Embedded image.
)

// Element is a single piece of report content in document order.
type Element struct {
	Kind     ElementKind
	Content  string // For prose: Markdown text. For display: the output data.
	Mime     string // MIME type for display elements.
	CellName string // Cell that produced this element (for display/error).
	Line     int    // Markdown line for error reporting.
	ImageRef string // For images: the resolved path or data URI.
}

// NewDocument creates an empty document.
func NewDocument() *Document {
	return &Document{}
}

// AddProse appends a prose element.
func (d *Document) AddProse(content string) {
	d.Elements = append(d.Elements, Element{Kind: KindProse, Content: content})
}

// AddDisplay appends a cell display output element.
func (d *Document) AddDisplay(cellName, mime, data string) {
	d.Elements = append(d.Elements, Element{
		Kind: KindDisplay, CellName: cellName, Mime: mime, Content: data,
	})
}

// AddError appends an error block at a cell's position.
func (d *Document) AddError(cellName string, line int, err error) {
	d.Elements = append(d.Elements, Element{
		Kind: KindError, CellName: cellName, Line: line, Content: err.Error(),
	})
}

// AddImage appends an embedded image element.
func (d *Document) AddImage(ref, mime string, data string) {
	d.Elements = append(d.Elements, Element{
		Kind: KindImage, ImageRef: ref, Mime: mime, Content: data,
	})
}

// Len returns the number of elements.
func (d *Document) Len() int {
	return len(d.Elements)
}
