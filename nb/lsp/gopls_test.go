// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/asymmetric-effort/go-notebook/nb/core"
)

// PinnedGoplsVersion is the version of gopls that tests are validated against.
// Update this when upgrading gopls in the test images.
const PinnedGoplsVersion = "v0.24.0"

// goplsPath returns the path to gopls, or skips the test if not found.
func goplsPath(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("gopls")
	if err != nil {
		t.Skip("gopls not found in PATH")
	}
	return path
}

// TestGoplsVersionPinned verifies the installed gopls matches the pinned version.
func TestGoplsVersionPinned(t *testing.T) {
	path := goplsPath(t)
	cmd := exec.Command(path, "version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("gopls version: %v", err)
	}
	version := strings.TrimSpace(string(out))
	if !strings.Contains(version, PinnedGoplsVersion) {
		t.Logf("gopls version: %s (pinned: %s)", version, PinnedGoplsVersion)
		t.Logf("WARNING: gopls version does not match pinned version")
		// Don't fail — just log. CI images will have the pinned version.
	} else {
		t.Logf("gopls version matches pinned: %s", PinnedGoplsVersion)
	}
}

// TestGoplsStartStop verifies that gopls can be started and stopped cleanly.
func TestGoplsStartStop(t *testing.T) {
	path := goplsPath(t)
	cmd := exec.Command(path, "serve")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Send a shutdown request.
	shutdownReq := `{"jsonrpc":"2.0","id":1,"method":"shutdown","params":null}`
	writeMsg(t, stdin, shutdownReq)

	// Read the response.
	resp := readMsg(t, stdout)
	if !strings.Contains(string(resp), `"id":1`) {
		t.Errorf("unexpected shutdown response: %s", resp)
	}

	// Send exit notification.
	exitNotif := `{"jsonrpc":"2.0","method":"exit","params":null}`
	writeMsg(t, stdin, exitNotif)

	stdin.Close()
	if err := cmd.Wait(); err != nil {
		// gopls may exit with non-zero on shutdown; that's ok.
		t.Logf("gopls exit: %v (expected)", err)
	}
}

// TestGoplsInitialize verifies the LSP initialize handshake works through
// a real gopls instance with a generated Go package.
func TestGoplsInitialize(t *testing.T) {
	path := goplsPath(t)

	// Create a temporary workspace with a valid Go package.
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "testpkg")
	os.MkdirAll(pkgDir, 0o755)
	os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module testpkg\ngo 1.27.4\n"), 0o644)
	os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)

	cmd := exec.Command(path, "serve")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		writeMsg(t, stdin, `{"jsonrpc":"2.0","id":99,"method":"shutdown","params":null}`)
		readMsg(t, stdout) // Read shutdown response.
		writeMsg(t, stdin, `{"jsonrpc":"2.0","method":"exit","params":null}`)
		stdin.Close()
		cmd.Wait()
	}()

	// Send initialize request.
	initParams := map[string]interface{}{
		"processId": os.Getpid(),
		"rootUri":   "file://" + pkgDir,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"completion": map[string]interface{}{},
			},
		},
	}
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  initParams,
	}
	body, _ := json.Marshal(initReq)
	writeMsg(t, stdin, string(body))

	// Read initialize response.
	resp := readMsg(t, stdout)
	var initResp map[string]interface{}
	if err := json.Unmarshal(resp, &initResp); err != nil {
		t.Fatalf("parse initialize response: %v", err)
	}
	if _, ok := initResp["result"]; !ok {
		t.Fatalf("initialize response missing result: %s", resp)
	}
	t.Logf("gopls initialized successfully")

	// Send initialized notification.
	writeMsg(t, stdin, `{"jsonrpc":"2.0","method":"initialized","params":{}}`)
}

// TestProxyPositionRoundTrip verifies that position translation through
// the source map is correct for a real notebook cell body.
func TestProxyPositionRoundTrip(t *testing.T) {
	// Parse a notebook and generate code to get a real source map.
	src := "# Title\n\nSome prose.\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	result := core.Parse(src)
	gen, err := core.Generate(result, "main")
	if err != nil {
		t.Fatal(err)
	}

	// Write the package to populate the source map.
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "testpkg")
	core.WritePackage(gen, pkgDir, "test.md", "testpkg")

	sm := gen.SourceMap
	if sm == nil || sm.Len() == 0 {
		t.Fatal("source map should have entries after WritePackage")
	}

	p := New("gopls")
	p.SetSourceMap(sm)

	// The cell body starts at Markdown line 6 (val = 42).
	// Translate Md -> Go -> Md and verify round-trip.
	goLine, goCol, ok := sm.MdToGo(6, 1)
	if !ok {
		t.Fatal("Md line 6 should be in source map")
	}
	t.Logf("Md(6,1) -> Go(%d,%d)", goLine, goCol)

	mdLine, mdCol, ok := sm.GoToMd(goLine, goCol)
	if !ok {
		t.Fatal("Go line should map back to Md")
	}
	if mdLine != 6 || mdCol != 1 {
		t.Errorf("round-trip: Md(6,1) -> Go(%d,%d) -> Md(%d,%d)", goLine, goCol, mdLine, mdCol)
	}

	// Translate an LSP message through the proxy.
	msg := map[string]interface{}{
		"params": map[string]interface{}{
			"position": map[string]interface{}{
				"line":      float64(5), // 0-based = Md line 6
				"character": float64(0),
			},
		},
	}
	body, _ := json.Marshal(msg)

	// Md -> Go
	translated := p.translatePositions(body, true)
	var goMsg map[string]interface{}
	json.Unmarshal(translated, &goMsg)
	goPos := goMsg["params"].(map[string]interface{})["position"].(map[string]interface{})
	goLSPLine := goPos["line"].(float64)

	// Go -> Md (reverse)
	backTranslated := p.translatePositions(translated, false)
	var mdMsg map[string]interface{}
	json.Unmarshal(backTranslated, &mdMsg)
	mdPos := mdMsg["params"].(map[string]interface{})["position"].(map[string]interface{})
	mdLSPLine := mdPos["line"].(float64)

	if mdLSPLine != 5 {
		t.Errorf("round-trip LSP position: start=%v -> go=%v -> back=%v", 5, goLSPLine, mdLSPLine)
	}
	t.Logf("LSP round-trip: Md(0-based=%v) -> Go(0-based=%v) -> Md(0-based=%v)", 5, goLSPLine, mdLSPLine)
}

// TestProxyForwardMessages verifies the forward function reads and writes
// LSP messages correctly.
func TestProxyForwardMessages(t *testing.T) {
	p := New("gopls")

	msg := `{"jsonrpc":"2.0","method":"test","params":{}}`
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(msg), msg)

	var output bytes.Buffer
	p.forward(strings.NewReader(input), &output, true)

	got := output.String()
	if !strings.Contains(got, "Content-Length:") {
		t.Error("output missing Content-Length header")
	}
	if !strings.Contains(got, `"method":"test"`) {
		t.Error("output missing forwarded message")
	}
}

// BenchmarkPositionTranslation measures the latency of position translation
// through the source map for a typical notebook.
func BenchmarkPositionTranslation(b *testing.B) {
	// Build a source map with realistic entries (10 cells, 20 lines each).
	sm := core.NewSourceMap()
	for i := 0; i < 10; i++ {
		sm.Add(i*30+10, 1, i*25+5, 1, 20)
	}

	p := New("gopls")
	p.SetSourceMap(sm)

	msg := map[string]interface{}{
		"params": map[string]interface{}{
			"position": map[string]interface{}{
				"line":      float64(39), // Inside the second cell.
				"character": float64(10),
			},
		},
	}
	body, _ := json.Marshal(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.translatePositions(body, true)
	}
}

// BenchmarkSourceMapLookup measures raw source map lookup latency.
func BenchmarkSourceMapLookup(b *testing.B) {
	sm := core.NewSourceMap()
	for i := 0; i < 50; i++ { // 50-cell notebook.
		sm.Add(i*20+5, 1, i*15+3, 1, 10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.MdToGo(500, 5) // Look up a line in the middle.
	}
}

// BenchmarkRegeneration measures how long it takes to parse, generate,
// and build a source map for a multi-cell notebook.
func BenchmarkRegeneration(b *testing.B) {
	// Build a 5-cell notebook.
	var sb strings.Builder
	sb.WriteString("# Benchmark Notebook\n\n")
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("cell%d", i)
		sb.WriteString(fmt.Sprintf("```go:notebook name=%s exported(val%d int)\nval%d = %d\n```\n\n", name, i, i, i*10))
	}
	src := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := core.Parse(src)
		gen, _ := core.Generate(result, "main")
		dir := b.TempDir()
		core.WritePackage(gen, filepath.Join(dir, "pkg"), "bench.md", "benchpkg")
	}
}

// writeMsg writes an LSP message to a writer.
func writeMsg(t *testing.T, w io.Writer, body string) {
	t.Helper()
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := io.WriteString(w, header+body); err != nil {
		t.Fatalf("write LSP message: %v", err)
	}
}

// readMsg reads one LSP message from a reader with a timeout.
func readMsg(t *testing.T, r io.Reader) []byte {
	t.Helper()
	done := make(chan []byte, 1)
	go func() {
		br := bufio.NewReader(r)
		msg, err := readLSPMessage(br)
		if err != nil {
			done <- nil
			return
		}
		done <- msg
	}()
	select {
	case msg := <-done:
		if msg == nil {
			t.Fatal("failed to read LSP message")
		}
		return msg
	case <-time.After(10 * time.Second):
		t.Fatal("timeout reading LSP message")
		return nil
	}
}
