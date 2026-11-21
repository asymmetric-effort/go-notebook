// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/asymmetric-effort/go-notebook/nb/core"
)

func TestReadLSPMessage(t *testing.T) {
	input := "Content-Length: 13\r\n\r\n{\"test\":true}"
	r := bufio.NewReader(strings.NewReader(input))
	msg, err := readLSPMessage(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(msg) != "{\"test\":true}" {
		t.Errorf("msg = %q", msg)
	}
}

func TestReadLSPMessageMissingLength(t *testing.T) {
	input := "Content-Type: text/plain\r\n\r\n{}"
	r := bufio.NewReader(strings.NewReader(input))
	_, err := readLSPMessage(r)
	if err == nil {
		t.Error("expected error for missing Content-Length")
	}
}

func TestWriteLSPMessage(t *testing.T) {
	var buf bytes.Buffer
	body := []byte("{\"test\":true}")
	err := writeLSPMessage(&buf, body)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "Content-Length: 13") {
		t.Errorf("missing Content-Length: %q", got)
	}
	if !strings.Contains(got, "{\"test\":true}") {
		t.Errorf("missing body: %q", got)
	}
}

func TestTranslatePositions(t *testing.T) {
	sm := core.NewSourceMap()
	sm.Add(10, 1, 5, 1, 3) // Md line 10 -> Go line 5, 3 lines.

	p := New("gopls")
	p.SetSourceMap(sm)

	// A message with a position at Markdown line 11 (0-based: 10).
	msg := map[string]interface{}{
		"params": map[string]interface{}{
			"position": map[string]interface{}{
				"line":      float64(10), // 0-based = Md line 11
				"character": float64(5),
			},
		},
	}
	body, _ := json.Marshal(msg)

	translated := p.translatePositions(body, true) // Md -> Go.

	var result map[string]interface{}
	json.Unmarshal(translated, &result)

	params := result["params"].(map[string]interface{})
	pos := params["position"].(map[string]interface{})
	line := pos["line"].(float64)
	// Md line 11 (1-based) -> Go line 6 (1-based) -> 0-based: 5.
	if line != 5 {
		t.Errorf("translated line = %v, want 5", line)
	}
}

func TestTranslatePositionsNoSourceMap(t *testing.T) {
	p := New("gopls")
	msg := []byte(`{"test": true}`)
	got := p.translatePositions(msg, true)
	if string(got) != string(msg) {
		t.Errorf("should pass through without source map: %q", got)
	}
}

func TestTranslatePositionsInvalidJSON(t *testing.T) {
	sm := core.NewSourceMap()
	sm.Add(1, 1, 1, 1, 1)
	p := New("gopls")
	p.SetSourceMap(sm)

	msg := []byte("not json")
	got := p.translatePositions(msg, true)
	if string(got) != "not json" {
		t.Errorf("should pass through invalid JSON: %q", got)
	}
}

func TestIsPosition(t *testing.T) {
	pos := map[string]interface{}{"line": float64(1), "character": float64(2)}
	if !isPosition(pos) {
		t.Error("should be position")
	}
	notPos := map[string]interface{}{"line": float64(1)}
	if isPosition(notPos) {
		t.Error("should not be position")
	}
}

func TestIsRange(t *testing.T) {
	r := map[string]interface{}{
		"start": map[string]interface{}{"line": float64(0), "character": float64(0)},
		"end":   map[string]interface{}{"line": float64(1), "character": float64(0)},
	}
	if !isRange(r) {
		t.Error("should be range")
	}
}

func TestTranslateRange(t *testing.T) {
	sm := core.NewSourceMap()
	sm.Add(10, 1, 5, 1, 5) // Md 10-14 -> Go 5-9

	p := New("gopls")
	p.SetSourceMap(sm)

	msg := map[string]interface{}{
		"result": map[string]interface{}{
			"range": map[string]interface{}{
				"start": map[string]interface{}{"line": float64(4), "character": float64(0)},
				"end":   map[string]interface{}{"line": float64(5), "character": float64(0)},
			},
		},
	}
	body, _ := json.Marshal(msg)

	// Go -> Md translation.
	translated := p.translatePositions(body, false)

	var result map[string]interface{}
	json.Unmarshal(translated, &result)

	r := result["result"].(map[string]interface{})
	rng := r["range"].(map[string]interface{})
	start := rng["start"].(map[string]interface{})
	startLine := start["line"].(float64)
	// Go line 5 (1-based) -> Md line 10 (1-based) -> 0-based: 9.
	if startLine != 9 {
		t.Errorf("start line = %v, want 9", startLine)
	}
}

func TestTranslateArray(t *testing.T) {
	sm := core.NewSourceMap()
	sm.Add(10, 1, 5, 1, 3)

	p := New("gopls")
	p.SetSourceMap(sm)

	msg := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"position": map[string]interface{}{
					"line":      float64(10),
					"character": float64(0),
				},
			},
		},
	}
	body, _ := json.Marshal(msg)

	translated := p.translatePositions(body, true)
	if string(translated) == string(body) {
		// If source map covers the line, it should change.
		var result map[string]interface{}
		json.Unmarshal(translated, &result)
		items := result["items"].([]interface{})
		item := items[0].(map[string]interface{})
		pos := item["position"].(map[string]interface{})
		line := pos["line"].(float64)
		if line != 5 {
			t.Errorf("array item line = %v, want 5", line)
		}
	}
}

func TestNew(t *testing.T) {
	p := New("/usr/bin/gopls")
	if p.goplsPath != "/usr/bin/gopls" {
		t.Errorf("goplsPath = %q", p.goplsPath)
	}
	if p.sourceMap == nil {
		t.Error("sourceMap should be initialized")
	}
}
