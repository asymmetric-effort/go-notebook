// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package lsp implements a stdio LSP proxy between the editor and gopls.
// It rewrites positions using the source map so the editor sees Markdown
// line numbers while gopls works on the generated Go source.
package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/asymmetric-effort/go-notebook/nb/core"
)

// Proxy is an LSP stdio proxy between the editor and gopls.
type Proxy struct {
	goplsPath string
	sourceMap *core.SourceMap
	mu        sync.RWMutex
}

// New creates a new LSP proxy. The goplsPath is the path to the gopls binary.
func New(goplsPath string) *Proxy {
	return &Proxy{
		goplsPath: goplsPath,
		sourceMap: core.NewSourceMap(),
	}
}

// SetSourceMap updates the source map used for position translation.
func (p *Proxy) SetSourceMap(sm *core.SourceMap) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sourceMap = sm
}

// Start spawns gopls and proxies messages between stdin/stdout and gopls.
// It blocks until either side closes.
func (p *Proxy) Start() error {
	cmd := exec.Command(p.goplsPath, "serve")
	cmd.Stderr = os.Stderr

	goplsIn, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("gopls stdin pipe: %w", err)
	}
	goplsOut, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("gopls stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start gopls: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Editor -> gopls: translate Markdown positions to Go positions.
	go func() {
		defer wg.Done()
		p.forward(os.Stdin, goplsIn, true)
		goplsIn.Close()
	}()

	// gopls -> editor: translate Go positions back to Markdown positions.
	go func() {
		defer wg.Done()
		p.forward(goplsOut, os.Stdout, false)
	}()

	wg.Wait()
	return cmd.Wait()
}

// forward reads LSP messages from src and writes them to dst,
// translating positions if needed. If toGo is true, positions are
// translated from Markdown to Go; otherwise from Go to Markdown.
func (p *Proxy) forward(src io.Reader, dst io.Writer, toGo bool) {
	reader := bufio.NewReader(src)
	for {
		msg, err := readLSPMessage(reader)
		if err != nil {
			return
		}

		// Translate positions in the message.
		translated := p.translatePositions(msg, toGo)

		if err := writeLSPMessage(dst, translated); err != nil {
			return
		}
	}
}

// readLSPMessage reads one LSP message (header + body) from a reader.
func readLSPMessage(r *bufio.Reader) ([]byte, error) {
	var contentLength int
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // End of headers.
		}
		if strings.HasPrefix(line, "Content-Length: ") {
			n, err := strconv.Atoi(strings.TrimPrefix(line, "Content-Length: "))
			if err != nil {
				return nil, fmt.Errorf("parse content length: %w", err)
			}
			contentLength = n
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

// writeLSPMessage writes one LSP message with headers to a writer.
func writeLSPMessage(w io.Writer, body []byte) error {
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := io.WriteString(w, header); err != nil {
		return err
	}
	_, err := w.Write(body)
	return err
}

// translatePositions rewrites line/character positions in an LSP JSON message.
// This is a best-effort translation: it looks for "position" and "range"
// objects and translates their line numbers.
func (p *Proxy) translatePositions(msg []byte, toGo bool) []byte {
	p.mu.RLock()
	sm := p.sourceMap
	p.mu.RUnlock()

	if sm == nil || sm.Len() == 0 {
		return msg
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(msg, &obj); err != nil {
		return msg // Not JSON, pass through.
	}

	translateObj(obj, sm, toGo)

	out, err := json.Marshal(obj)
	if err != nil {
		return msg
	}
	return out
}

// translateObj recursively walks a JSON object and translates positions.
func translateObj(obj map[string]interface{}, sm *core.SourceMap, toGo bool) {
	for key, val := range obj {
		switch v := val.(type) {
		case map[string]interface{}:
			// Check if this looks like a Position or Range.
			if isPosition(v) {
				translatePosition(v, sm, toGo)
			} else if isRange(v) {
				if start, ok := v["start"].(map[string]interface{}); ok {
					translatePosition(start, sm, toGo)
				}
				if end, ok := v["end"].(map[string]interface{}); ok {
					translatePosition(end, sm, toGo)
				}
			} else {
				translateObj(v, sm, toGo)
			}
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					translateObj(m, sm, toGo)
				}
			}
		}
		_ = key
	}
}

func isPosition(obj map[string]interface{}) bool {
	_, hasLine := obj["line"]
	_, hasChar := obj["character"]
	return hasLine && hasChar && len(obj) == 2
}

func isRange(obj map[string]interface{}) bool {
	_, hasStart := obj["start"]
	_, hasEnd := obj["end"]
	return hasStart && hasEnd
}

func translatePosition(pos map[string]interface{}, sm *core.SourceMap, toGo bool) {
	lineVal, ok := pos["line"].(float64)
	if !ok {
		return
	}
	charVal, ok := pos["character"].(float64)
	if !ok {
		return
	}

	line := int(lineVal) + 1 // LSP is 0-based, source map is 1-based.
	col := int(charVal) + 1

	var newLine, newCol int
	var found bool
	if toGo {
		newLine, newCol, found = sm.MdToGo(line, col)
	} else {
		newLine, newCol, found = sm.GoToMd(line, col)
	}

	if found {
		pos["line"] = float64(newLine - 1) // Back to 0-based.
		pos["character"] = float64(newCol - 1)
	}
}
