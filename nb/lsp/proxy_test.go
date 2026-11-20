// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package lsp

import (
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/core"
)

func TestNewProxy(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
}

func TestTranslateWithSourceMap(t *testing.T) {
	p := New()
	sm := core.NewSourceMap()
	sm.Add(10, 1, 20, 1, 5)
	p.UpdateSourceMap(sm)

	goLine, _, ok := p.TranslateMdToGo(12, 1)
	if !ok || goLine != 22 {
		t.Errorf("MdToGo(12) = %d, %v; want 22, true", goLine, ok)
	}

	mdLine, _, ok := p.TranslateGoToMd(22, 1)
	if !ok || mdLine != 12 {
		t.Errorf("GoToMd(22) = %d, %v; want 12, true", mdLine, ok)
	}
}

func TestTranslateNoMapping(t *testing.T) {
	p := New()
	_, _, ok := p.TranslateMdToGo(100, 1)
	if ok {
		t.Error("expected no mapping")
	}
}

func TestCloseNilProcess(t *testing.T) {
	p := New()
	if err := p.Close(); err != nil {
		t.Errorf("Close() = %v", err)
	}
}
