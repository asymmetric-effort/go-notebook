// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package document

import (
	"errors"
	"testing"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument()
	if doc.Len() != 0 {
		t.Errorf("Len() = %d, want 0", doc.Len())
	}
}

func TestAddProse(t *testing.T) {
	doc := NewDocument()
	doc.AddProse("# Hello")
	if doc.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", doc.Len())
	}
	if doc.Elements[0].Kind != KindProse {
		t.Errorf("Kind = %v, want Prose", doc.Elements[0].Kind)
	}
	if doc.Elements[0].Content != "# Hello" {
		t.Errorf("Content = %q", doc.Elements[0].Content)
	}
}

func TestAddDisplay(t *testing.T) {
	doc := NewDocument()
	doc.AddDisplay("test", "text/markdown", "output data")
	if doc.Len() != 1 {
		t.Fatalf("Len() = %d", doc.Len())
	}
	e := doc.Elements[0]
	if e.Kind != KindDisplay || e.CellName != "test" || e.Mime != "text/markdown" {
		t.Errorf("element = %+v", e)
	}
}

func TestAddError(t *testing.T) {
	doc := NewDocument()
	doc.AddError("load", 42, errors.New("file not found"))
	if doc.Len() != 1 {
		t.Fatalf("Len() = %d", doc.Len())
	}
	e := doc.Elements[0]
	if e.Kind != KindError || e.CellName != "load" || e.Line != 42 {
		t.Errorf("element = %+v", e)
	}
}

func TestAddImage(t *testing.T) {
	doc := NewDocument()
	doc.AddImage("logo.png", "image/png", "base64data")
	if doc.Len() != 1 {
		t.Fatalf("Len() = %d", doc.Len())
	}
	e := doc.Elements[0]
	if e.Kind != KindImage || e.ImageRef != "logo.png" {
		t.Errorf("element = %+v", e)
	}
}

func TestDocumentOrder(t *testing.T) {
	doc := NewDocument()
	doc.AddProse("# Title")
	doc.AddDisplay("load", "text/markdown", "data")
	doc.AddProse("Between")
	doc.AddError("calc", 10, errors.New("fail"))
	if doc.Len() != 4 {
		t.Fatalf("Len() = %d, want 4", doc.Len())
	}
	kinds := []ElementKind{KindProse, KindDisplay, KindProse, KindError}
	for i, want := range kinds {
		if doc.Elements[i].Kind != want {
			t.Errorf("Elements[%d].Kind = %v, want %v", i, doc.Elements[i].Kind, want)
		}
	}
}
