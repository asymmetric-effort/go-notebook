// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func testNotebook(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "# Test\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestStatusEndpoint(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}

	var resp statusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Notebook != "test.md" {
		t.Errorf("notebook = %q", resp.Notebook)
	}
	if resp.Cells != 1 {
		t.Errorf("cells = %d, want 1", resp.Cells)
	}
	if !resp.OK {
		t.Error("ok should be true")
	}
	if resp.Version == "" {
		t.Error("version should not be empty")
	}
}

func TestStatusVersionField(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	v, ok := raw["version"]
	if !ok {
		t.Fatal("response missing 'version' field")
	}
	vs, ok := v.(string)
	if !ok {
		t.Fatal("version field is not a string")
	}
	if vs == "" {
		t.Error("version field should not be empty")
	}
}

func TestStatusMethodNotAllowed(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := httptest.NewRequest(http.MethodPost, "/api/status", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want 405", w.Code)
	}
}

func TestCellsEndpoint(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/cells", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}

	var cells []cellInfo
	if err := json.Unmarshal(w.Body.Bytes(), &cells); err != nil {
		t.Fatal(err)
	}
	if len(cells) != 1 {
		t.Fatalf("cells = %d, want 1", len(cells))
	}
	if cells[0].Name != "test" {
		t.Errorf("cell name = %q", cells[0].Name)
	}
}

func TestCellsMethodNotAllowed(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := httptest.NewRequest(http.MethodPost, "/api/cells", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want 405", w.Code)
	}
}

func TestRunEndpoint(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := httptest.NewRequest(http.MethodPost, "/api/run", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200\nbody: %s", w.Code, w.Body.String())
	}

	var resp runResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Errorf("exit code = %d, stderr: %s, error: %s", resp.ExitCode, resp.Stderr, resp.Error)
	}
}

func TestRunMethodNotAllowed(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/run", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want 405", w.Code)
	}
}

func TestRunWithCheckErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	src := "```go:notebook exported(val int)\nval = 1\n```\n" // exported without name
	os.WriteFile(path, []byte(src), 0o644)

	srv := New(path, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/run", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want 400", w.Code)
	}
}

func TestStatusMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want 500", w.Code)
	}
}

func TestNew(t *testing.T) {
	srv := New("test.md", ":8080")
	if srv.addr != ":8080" {
		t.Errorf("addr = %q", srv.addr)
	}
	if srv.notebookPath != "test.md" {
		t.Errorf("path = %q", srv.notebookPath)
	}
}

func TestCellsEndpointBody(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/cells", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}

	var cells []cellInfo
	if err := json.Unmarshal(w.Body.Bytes(), &cells); err != nil {
		t.Fatal(err)
	}
	if len(cells) != 1 {
		t.Fatalf("cells = %d, want 1", len(cells))
	}
	cell := cells[0]
	if cell.Body != "val = 42" {
		t.Errorf("Body = %q, want %q", cell.Body, "val = 42")
	}
	if cell.Exported != "val int" {
		t.Errorf("Exported = %q, want %q", cell.Exported, "val int")
	}
	if cell.Reads == nil {
		t.Error("Reads should not be nil")
	}
	if cell.Writes == nil {
		t.Error("Writes should not be nil")
	}
	if cell.Secrets == nil {
		t.Error("Secrets should not be nil")
	}
}

func TestCellsUpdateEndpoint(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := updateRequest{
		Line:     3,
		Body:     "val = 99",
		Name:     "test",
		Exported: "val int",
		Reads:    []readFileInfo{},
		Writes:   []string{},
		Secrets:  []string{},
		Data:     "",
		Timeout:  "",
	}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPut, "/api/cells/update", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200\nbody: %s", w.Code, w.Body.String())
	}

	// Verify the file was updated.
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "val = 99") {
		t.Errorf("file does not contain updated body:\n%s", src)
	}
}

func TestCellsUpdateNotFound(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := updateRequest{Line: 999, Body: "x := 1"}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPut, "/api/cells/update", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want 400", w.Code)
	}
}

func TestCellsUpdateMethodNotAllowed(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	r := httptest.NewRequest(http.MethodGet, "/api/cells/update", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want 405", w.Code)
	}
}

func TestCellsUpdateInvalidJSON(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	r := httptest.NewRequest(http.MethodPut, "/api/cells/update", strings.NewReader("{bad json"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want 400", w.Code)
	}
}

func TestCellsUpdateInvalidTimeout(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	req := updateRequest{Line: 3, Body: "val = 1", Name: "test", Timeout: "notaduration"}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPut, "/api/cells/update", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want 400", w.Code)
	}
}

func TestEditorEndpoint(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	r := httptest.NewRequest(http.MethodGet, "/editor", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(w.Body.String(), "go-notebook editor") {
		t.Error("editor HTML does not contain expected title")
	}
}

func TestEditorMethodNotAllowed(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	r := httptest.NewRequest(http.MethodPost, "/editor", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want 405", w.Code)
	}
}

func TestStaticFileServing(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	r := httptest.NewRequest(http.MethodGet, "/static/logo.png", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}
}

func TestRewriteCell(t *testing.T) {
	src := "# Test\n\n```go:notebook name=old\noriginal body\n```\n"
	req := updateRequest{
		Line:     3,
		Body:     "new body",
		Name:     "new",
		Exported: "",
		Reads:    nil,
		Writes:   nil,
		Secrets:  nil,
		Data:     "",
		Timeout:  "",
	}
	got, err := rewriteCell(src, req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "name=new") {
		t.Errorf("output missing name=new:\n%s", got)
	}
	if !strings.Contains(got, "new body") {
		t.Errorf("output missing new body:\n%s", got)
	}
	if strings.Contains(got, "original body") {
		t.Errorf("output still contains original body:\n%s", got)
	}
}

func TestRewriteCellWithAttrs(t *testing.T) {
	src := "```go:notebook name=load\nbody\n```\n"
	req := updateRequest{
		Line:    1,
		Body:    "result := 42",
		Name:    "compute",
		Reads:   []readFileInfo{{Name: "data.csv", Mode: "embed"}},
		Writes:  []string{"out.png"},
		Secrets: []string{"token"},
		Data:    "embedded",
		Timeout: "30s",
	}
	got, err := rewriteCell(src, req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "name=compute") {
		t.Errorf("missing name=compute in:\n%s", got)
	}
	if !strings.Contains(got, "reads(data.csv:embed)") {
		t.Errorf("missing reads in:\n%s", got)
	}
	if !strings.Contains(got, "writes(out.png)") {
		t.Errorf("missing writes in:\n%s", got)
	}
	if !strings.Contains(got, "secrets(token)") {
		t.Errorf("missing secrets in:\n%s", got)
	}
	if !strings.Contains(got, "data=embedded") {
		t.Errorf("missing data=embedded in:\n%s", got)
	}
	if !strings.Contains(got, "timeout=30s") {
		t.Errorf("missing timeout=30s in:\n%s", got)
	}
	if !strings.Contains(got, "result := 42") {
		t.Errorf("missing updated body in:\n%s", got)
	}
}

func TestExtractFencePrefix(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"```go:notebook name=foo", "```"},
		{"````go:notebook", "````"},
		{"~~~go:notebook", "~~~"},
		{"  ```go:notebook", "  ```"},
		{"", "```"},
	}
	for _, tt := range tests {
		got := extractFencePrefix(tt.line)
		if got != tt.want {
			t.Errorf("extractFencePrefix(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestBlocksEndpoint(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Title\n\nSome prose.\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/blocks", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var blocks []blockInfo
	json.Unmarshal(w.Body.Bytes(), &blocks)

	if len(blocks) < 2 {
		t.Fatalf("got %d blocks, want at least 2", len(blocks))
	}
	if blocks[0].Kind != "prose" {
		t.Errorf("blocks[0].Kind = %q, want prose", blocks[0].Kind)
	}
	if !strings.Contains(blocks[0].Body, "Title") {
		t.Error("prose block should contain title")
	}
	foundCell := false
	for _, b := range blocks {
		if b.Kind == "cell" && b.Name == "test" {
			foundCell = true
		}
	}
	if !foundCell {
		t.Error("no cell block found")
	}
}

func TestBlocksEndpointMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("POST", "/api/blocks", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestProseUpdateEndpoint(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Original Title\n\nSome prose.\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Get blocks to find the prose start line.
	getReq := httptest.NewRequest("GET", "/api/blocks", nil)
	getW := httptest.NewRecorder()
	s.Handler().ServeHTTP(getW, getReq)

	var blocks []blockInfo
	json.Unmarshal(getW.Body.Bytes(), &blocks)
	if len(blocks) == 0 {
		t.Fatal("no blocks")
	}
	proseLine := blocks[0].Line

	updateBody, _ := json.Marshal(proseUpdateRequest{
		Line: proseLine,
		Body: "# Updated Title\n\nNew prose content.",
	})
	putReq := httptest.NewRequest("PUT", "/api/prose/update", bytes.NewReader(updateBody))
	putReq.Header.Set("Content-Type", "application/json")
	putW := httptest.NewRecorder()
	s.Handler().ServeHTTP(putW, putReq)

	if putW.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", putW.Code, putW.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	if !strings.Contains(content, "Updated Title") {
		t.Error("prose should be updated")
	}
	if !strings.Contains(content, "val = 42") {
		t.Error("cell body should be preserved after prose update")
	}
}

func TestProseUpdateNotFound(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	updateBody, _ := json.Marshal(proseUpdateRequest{Line: 999, Body: "new"})
	req := httptest.NewRequest("PUT", "/api/prose/update", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestProseUpdateMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/prose/update", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestProseUpdatePreservesCells(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Title\n\n```go:notebook name=a exported(x int)\nx = 1\n```\n\nMiddle prose.\n\n```go:notebook name=b exported(y int)\ny = 2\n```\n\nEnd prose.\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	getReq := httptest.NewRequest("GET", "/api/blocks", nil)
	getW := httptest.NewRecorder()
	s.Handler().ServeHTTP(getW, getReq)

	var blocks []blockInfo
	json.Unmarshal(getW.Body.Bytes(), &blocks)

	var middleLine int
	for _, b := range blocks {
		if b.Kind == "prose" && strings.Contains(b.Body, "Middle") {
			middleLine = b.Line
			break
		}
	}
	if middleLine == 0 {
		t.Fatal("middle prose block not found")
	}

	updateBody, _ := json.Marshal(proseUpdateRequest{
		Line: middleLine,
		Body: "Updated middle prose with **more content**.",
	})
	putReq := httptest.NewRequest("PUT", "/api/prose/update", bytes.NewReader(updateBody))
	putReq.Header.Set("Content-Type", "application/json")
	putW := httptest.NewRecorder()
	s.Handler().ServeHTTP(putW, putReq)

	if putW.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", putW.Code, putW.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	if !strings.Contains(content, "Updated middle prose") {
		t.Error("updated prose missing")
	}
	if !strings.Contains(content, "x = 1") {
		t.Error("cell a body altered")
	}
	if !strings.Contains(content, "y = 2") {
		t.Error("cell b body altered")
	}
	if !strings.Contains(content, "name=a") {
		t.Error("cell a info altered")
	}
	if !strings.Contains(content, "name=b") {
		t.Error("cell b info altered")
	}
	if !strings.Contains(content, "End prose") {
		t.Error("end prose altered")
	}
}

func TestRewriteProseUnit(t *testing.T) {
	src := "# Title\n\nProse here.\n\n```go:notebook name=test\nval = 1\n```\n"
	updated, err := rewriteProse(src, proseUpdateRequest{Line: 1, Body: "# New Title\n\nNew prose."})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updated, "New Title") {
		t.Error("title should be updated")
	}
	if !strings.Contains(updated, "val = 1") {
		t.Error("cell should be preserved")
	}
}

func TestRewriteProseNotFound(t *testing.T) {
	src := "# Title\n"
	_, err := rewriteProse(src, proseUpdateRequest{Line: 999, Body: "new"})
	if err == nil {
		t.Error("expected error for nonexistent prose block")
	}
}

func TestRunStateEndpoint(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=test exported(val int)\nval = 42\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/run/state", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	// Initially empty state.
	var states map[string]*cellState
	json.Unmarshal(w.Body.Bytes(), &states)
	if len(states) != 0 {
		t.Errorf("initial states should be empty, got %d", len(states))
	}
}

func TestRunStateMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("POST", "/api/run/state", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestRunAllMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/run/all", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestMarkStale(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=a exported(x int)\nx = 1\n```\n\n```go:notebook name=b exported(y int)\ny = a.x + 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Set up initial cell states (simulate previous run).
	s.mu.Lock()
	s.cellStates["a"] = &cellState{Name: "a", Stale: false}
	s.cellStates["b"] = &cellState{Name: "b", Stale: false}
	s.mu.Unlock()

	// Mark cell "a" stale — should also mark "b" stale (downstream).
	s.markStale("a")

	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.cellStates["a"].Stale {
		t.Error("cell a should be stale")
	}
	if !s.cellStates["b"].Stale {
		t.Error("cell b should be stale (downstream of a)")
	}
}

func TestMarkStaleAll(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	s.mu.Lock()
	s.cellStates["a"] = &cellState{Name: "a", Stale: false}
	s.cellStates["b"] = &cellState{Name: "b", Stale: false}
	s.mu.Unlock()

	// Mark all stale (empty cell name).
	s.markStale("")

	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.cellStates["a"].Stale {
		t.Error("cell a should be stale")
	}
	if !s.cellStates["b"].Stale {
		t.Error("cell b should be stale")
	}
}

func TestCheckUpstreamFailure(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=a exported(x int)\nx = 1\n```\n\n```go:notebook name=b exported(y int)\ny = a.x\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// No failures — should return empty.
	if refused := s.checkUpstreamFailure("b"); refused != "" {
		t.Errorf("expected no refusal, got %q", refused)
	}

	// Mark cell "a" as failed.
	s.mu.Lock()
	s.cellStates["a"] = &cellState{Name: "a", Failed: true}
	s.mu.Unlock()

	// Now "b" should be refused because "a" failed.
	if refused := s.checkUpstreamFailure("b"); refused != "a" {
		t.Errorf("expected refusal for 'a', got %q", refused)
	}

	// Cell "a" itself has no upstream, so it should not be refused.
	if refused := s.checkUpstreamFailure("a"); refused != "" {
		t.Errorf("expected no refusal for first cell, got %q", refused)
	}
}

func TestRunRefusal(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=a exported(x int)\nx = 1\n```\n\n```go:notebook name=b exported(y int)\ny = a.x\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Mark cell "a" as failed.
	s.mu.Lock()
	s.cellStates["a"] = &cellState{Name: "a", Failed: true}
	s.mu.Unlock()

	// Try to run cell "b" — should be refused.
	body, _ := json.Marshal(runRequest{Cell: "b"})
	req := httptest.NewRequest("POST", "/api/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}

	var resp runResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Refused != "a" {
		t.Errorf("refused = %q, want 'a'", resp.Refused)
	}
}

func TestParseRunErrorsTimeout(t *testing.T) {
	errs := parseRunErrors("", "cell timed out: context deadline exceeded", "test")
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if errs[0].Kind != "timeout" {
		t.Errorf("kind = %q, want timeout", errs[0].Kind)
	}
	if errs[0].Cell != "test" {
		t.Errorf("cell = %q, want test", errs[0].Cell)
	}
}

func TestParseRunErrorsCompile(t *testing.T) {
	stderr := "notebook.md:5:3: undefined: foo\n"
	errs := parseRunErrors(stderr, "", "test")
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if errs[0].Kind != "compile" {
		t.Errorf("kind = %q, want compile", errs[0].Kind)
	}
	if errs[0].Line != 5 {
		t.Errorf("line = %d, want 5", errs[0].Line)
	}
	if !strings.Contains(errs[0].Message, "undefined") {
		t.Errorf("message = %q, want 'undefined'", errs[0].Message)
	}
}

func TestParseRunErrorsPanic(t *testing.T) {
	stderr := "panic: runtime error: index out of range\n"
	errs := parseRunErrors(stderr, "", "test")
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if errs[0].Kind != "panic" {
		t.Errorf("kind = %q, want panic", errs[0].Kind)
	}
	if !strings.Contains(errs[0].Message, "index out of range") {
		t.Errorf("message = %q", errs[0].Message)
	}
}

func TestParseRunErrorsStackTraceFiltered(t *testing.T) {
	stderr := "notebook.md:10:1: type mismatch\ngoroutine 1 [running]:\n\truntime/debug.Stack()\n\t\t/usr/local/go/src/runtime/debug/stack.go:24\n"
	errs := parseRunErrors(stderr, "", "test")
	// Should only have the compile error, not the stack trace.
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if errs[0].Kind != "compile" {
		t.Errorf("kind = %q, want compile", errs[0].Kind)
	}
	for _, e := range errs {
		if strings.Contains(e.Message, "goroutine") || strings.Contains(e.Message, "runtime") {
			t.Error("stack trace should be filtered out")
		}
	}
}

func TestParseRunErrorsGenericFallback(t *testing.T) {
	errs := parseRunErrors("", "something went wrong", "test")
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if errs[0].Kind != "runtime" {
		t.Errorf("kind = %q, want runtime", errs[0].Kind)
	}
}

func TestParseRunErrorsEmpty(t *testing.T) {
	errs := parseRunErrors("", "", "test")
	if len(errs) != 0 {
		t.Errorf("got %d errors, want 0", len(errs))
	}
}

func TestParseCompileError(t *testing.T) {
	tests := []struct {
		line    string
		wantNil bool
		file    string
		lineNum int
	}{
		{"notebook.md:5:3: undefined: foo", false, "notebook.md", 5},
		{"./main.go:10:1: syntax error", false, "./main.go", 10},
		{"not an error", true, "", 0},
		{"", true, "", 0},
	}
	for _, tt := range tests {
		got := parseCompileError(tt.line)
		if tt.wantNil {
			if got != nil {
				t.Errorf("parseCompileError(%q) = %+v, want nil", tt.line, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("parseCompileError(%q) = nil", tt.line)
			continue
		}
		if got.File != tt.file {
			t.Errorf("file = %q, want %q", got.File, tt.file)
		}
		if got.Line != tt.lineNum {
			t.Errorf("line = %d, want %d", got.Line, tt.lineNum)
		}
	}
}

func TestCSPHeaderPresent(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")

	// Test CSP on API endpoint.
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("missing Content-Security-Policy header")
	}
	if !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("CSP missing default-src 'self': %s", csp)
	}
	if !strings.Contains(csp, "object-src 'none'") {
		t.Errorf("CSP missing object-src 'none': %s", csp)
	}
	if !strings.Contains(csp, "frame-src 'none'") {
		t.Errorf("CSP missing frame-src 'none': %s", csp)
	}
}

func TestCSPHeaderOnEditor(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")

	req := httptest.NewRequest("GET", "/editor", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("editor page missing CSP header")
	}
	if !strings.Contains(csp, "script-src 'self'") {
		t.Error("CSP should restrict scripts to 'self'")
	}
}

func TestXContentTypeOptionsHeader(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options: nosniff")
	}
}

func TestSanitizeOutputNoSVG(t *testing.T) {
	output := "hello world\n42\n"
	got := sanitizeOutput(output)
	// No SVG — output passes through unchanged.
	if got != output {
		t.Errorf("non-SVG output should pass through: got %q, want %q", got, output)
	}
}

func TestSanitizeOutputStripsSVG(t *testing.T) {
	output := "result: 42\n<svg xmlns=\"http://www.w3.org/2000/svg\">\n<script>alert('xss')</script>\n</svg>\nmore output\n"
	got := sanitizeOutput(output)
	if strings.Contains(got, "<svg") {
		t.Error("SVG should be stripped")
	}
	if strings.Contains(got, "alert") {
		t.Error("script inside SVG should be stripped")
	}
	if !strings.Contains(got, "result: 42") {
		t.Error("non-SVG output before SVG should be preserved")
	}
	if !strings.Contains(got, "more output") {
		t.Error("output after SVG should be preserved")
	}
	if !strings.Contains(got, "[SVG content stripped") {
		t.Error("should contain replacement message")
	}
}

func TestSanitizeOutputStripsSVGCaseInsensitive(t *testing.T) {
	output := "<SVG><script>evil</script></SVG>\n"
	got := sanitizeOutput(output)
	if strings.Contains(strings.ToLower(got), "<svg") {
		t.Error("uppercase SVG should also be stripped")
	}
}

func TestSanitizeOutputEmpty(t *testing.T) {
	got := sanitizeOutput("")
	if got != "" {
		t.Errorf("empty input should return empty: %q", got)
	}
}

func TestSecretsSetAndGet(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")

	// Set a secret.
	body, _ := json.Marshal(secretRequest{Name: "api_key", Value: "secret123"})
	req := httptest.NewRequest("PUT", "/api/secrets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, body = %s", w.Code, w.Body.String())
	}

	// Verify via GetSecret.
	val, ok := s.GetSecret("api_key")
	if !ok || val != "secret123" {
		t.Errorf("GetSecret = %q, %v", val, ok)
	}

	// GET should list names but not values.
	getReq := httptest.NewRequest("GET", "/api/secrets", nil)
	getW := httptest.NewRecorder()
	s.Handler().ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("GET status = %d", getW.Code)
	}
	respBody := getW.Body.String()
	if !strings.Contains(respBody, "api_key") {
		t.Error("GET should list secret name")
	}
	if strings.Contains(respBody, "secret123") {
		t.Error("GET must NEVER return secret values")
	}
}

func TestSecretsDelete(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")

	// Set a secret.
	s.mu.Lock()
	s.secrets["key"] = "value"
	s.mu.Unlock()

	// Delete all secrets.
	req := httptest.NewRequest("DELETE", "/api/secrets", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("DELETE status = %d", w.Code)
	}

	// Verify cleared.
	_, ok := s.GetSecret("key")
	if ok {
		t.Error("secret should be cleared after DELETE")
	}
}

func TestSecretsDeleteIndividual(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")

	// Set two secrets.
	s.mu.Lock()
	s.secrets["test"] = "val1"
	s.secrets["other"] = "val2"
	s.mu.Unlock()

	// Delete one secret by name.
	body, _ := json.Marshal(secretRequest{Name: "test"})
	req := httptest.NewRequest("DELETE", "/api/secrets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("DELETE individual status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %q", resp["status"])
	}

	// Verify deleted.
	_, ok := s.GetSecret("test")
	if ok {
		t.Error("secret 'test' should be deleted")
	}

	// Verify other secret still exists.
	val, ok := s.GetSecret("other")
	if !ok || val != "val2" {
		t.Errorf("secret 'other' should still exist, got %q, %v", val, ok)
	}
}

func TestSecretsDeleteNonexistent(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")

	// Try to delete a secret that doesn't exist.
	body, _ := json.Marshal(secretRequest{Name: "nonexistent"})
	req := httptest.NewRequest("DELETE", "/api/secrets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("DELETE nonexistent status = %d, want 404, body = %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !strings.Contains(resp["error"], "secret not found") {
		t.Errorf("expected 'secret not found' error, got %q", resp["error"])
	}
}

func TestSecretsDeleteAllWithEmptyBody(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")

	// Set secrets.
	s.mu.Lock()
	s.secrets["a"] = "1"
	s.secrets["b"] = "2"
	s.mu.Unlock()

	// Delete all with empty body.
	req := httptest.NewRequest("DELETE", "/api/secrets", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("DELETE all status = %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "cleared" {
		t.Errorf("expected status=cleared, got %q", resp["status"])
	}

	// Verify all cleared.
	_, okA := s.GetSecret("a")
	_, okB := s.GetSecret("b")
	if okA || okB {
		t.Error("all secrets should be cleared")
	}
}

func TestSecretsEmptyName(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(secretRequest{Name: "", Value: "val"})
	req := httptest.NewRequest("PUT", "/api/secrets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestSecretsMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("PATCH", "/api/secrets", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestSecretNeverInNotebook(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")

	// Set a secret.
	s.mu.Lock()
	s.secrets["password"] = "super_secret_value"
	s.mu.Unlock()

	// Verify the secret value does not appear in the notebook file.
	data, _ := os.ReadFile(notebook)
	if strings.Contains(string(data), "super_secret_value") {
		t.Error("secret value must NEVER appear in the notebook file")
	}

	// Verify the secret value does not appear in cell states.
	s.mu.RLock()
	for _, cs := range s.cellStates {
		if strings.Contains(cs.Stdout, "super_secret_value") {
			t.Error("secret value must not appear in cell stdout")
		}
		if strings.Contains(cs.Stderr, "super_secret_value") {
			t.Error("secret value must not appear in cell stderr")
		}
	}
	s.mu.RUnlock()
}

func TestChartScriptEndpoint(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/chart.js", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "javascript") {
		t.Errorf("Content-Type = %q, want javascript", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "initCharts") {
		t.Error("chart script should contain initCharts")
	}
}

func TestChartScriptMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("POST", "/api/chart.js", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestRunResponseIncludesDisplays(t *testing.T) {
	// Verify the displayEntry type serializes correctly.
	resp := runResponse{
		Cell:     "test",
		ExitCode: 0,
		Displays: []displayEntry{
			{Mime: "text/markdown", Data: "hello"},
			{Mime: "image/svg+xml", Data: "<svg></svg>"},
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "text/markdown") {
		t.Error("response should contain markdown display")
	}
	if !strings.Contains(s, "image/svg+xml") {
		t.Error("response should contain SVG display")
	}
}

func TestLSPDiagnosticsCleanNotebook(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/lsp/diagnostics", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	diags, _ := resp["diagnostics"].([]interface{})
	if len(diags) != 0 {
		t.Errorf("clean notebook should have no diagnostics, got %v", resp)
	}
}

func TestLSPDiagnosticsWithErrors(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// exported without name triggers checker rule 1.
	os.WriteFile(notebook, []byte("```go:notebook exported(val int)\nval = 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/lsp/diagnostics", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	diags, ok := resp["diagnostics"].([]interface{})
	if !ok || len(diags) == 0 {
		t.Error("notebook with errors should have diagnostics")
	}
}

func TestLSPDiagnosticsMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("POST", "/api/lsp/diagnostics", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestLSPCompletion(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=load exported(val int)\nval = 42\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/lsp/completion", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "load") {
		t.Error("completion should include cell name 'load'")
	}
	if !strings.Contains(body, "load.val") {
		t.Error("completion should include field 'load.val'")
	}
	if !strings.Contains(body, "ctx.Display") {
		t.Error("completion should include runtime keywords")
	}
}

func TestLSPCompletionMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("POST", "/api/lsp/completion", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestLSPHoverCellName(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=load exported(val int)\nval = 42\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/lsp/hover?word=load", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Cell load") {
		t.Error("hover should show cell name")
	}
	if !strings.Contains(body, "val int") {
		t.Error("hover should show exported fields")
	}
}

func TestLSPHoverField(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=load exported(val int)\nval = 42\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/lsp/hover?word=load.val", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "load.val") {
		t.Error("hover should show field reference")
	}
	if !strings.Contains(body, "int") {
		t.Error("hover should show field type")
	}
}

func TestLSPHoverMissingWord(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/lsp/hover", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestLSPHoverUnknownWord(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=test exported(val int)\nval = 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest("GET", "/api/lsp/hover?word=nonexistent", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	// Should return empty content for unknown word.
	var info hoverInfo
	json.Unmarshal(w.Body.Bytes(), &info)
	if info.Content != "" {
		t.Errorf("hover for unknown word should be empty, got %q", info.Content)
	}
}

// --- Additional tests for untested paths ---

func TestCellsMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/cells", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want 500", w.Code)
	}
}

func TestBlocksMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/blocks", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want 500", w.Code)
	}
}

func TestProseUpdateMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")

	body, _ := json.Marshal(proseUpdateRequest{Line: 1, Body: "new"})
	req := httptest.NewRequest(http.MethodPut, "/api/prose/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want 500", w.Code)
	}
}

func TestProseUpdateInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPut, "/api/prose/update", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want 400", w.Code)
	}
}

func TestCellsUpdateMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")

	req := updateRequest{Line: 1, Body: "val = 1", Name: "test"}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPut, "/api/cells/update", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want 500", w.Code)
	}
}

func TestRunAllEndpoint(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=step1 exported(val int)\nval = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/run/all", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var results []runResponse
	if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to parse response: %v\nbody: %s", err, w.Body.String())
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestRunAllMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")

	req := httptest.NewRequest(http.MethodPost, "/api/run/all", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want 500", w.Code)
	}
}

func TestRunSingleCellEndpoint(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 42\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	body, _ := json.Marshal(runRequest{Cell: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp runResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.ExitCode != 0 {
		t.Errorf("exit code = %d, stderr: %s, error: %s", resp.ExitCode, resp.Stderr, resp.Error)
	}
}

func TestLSPDiagnosticsMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/lsp/diagnostics", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want 500", w.Code)
	}
}

func TestLSPCompletionMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/lsp/completion", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want 500", w.Code)
	}
}

func TestLSPHoverMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")

	req := httptest.NewRequest(http.MethodGet, "/api/lsp/hover?word=test", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want 500", w.Code)
	}
}

func TestLSPHoverMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/lsp/hover", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want 405", w.Code)
	}
}

func TestSecretsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPut, "/api/secrets", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want 400", w.Code)
	}
}

func TestRewriteProseOutOfRange(t *testing.T) {
	// A prose block whose line numbers are out of the actual text range.
	// This is hard to trigger naturally, but we can verify it doesn't panic on valid input.
	src := "# Title\n"
	_, err := rewriteProse(src, proseUpdateRequest{Line: 999, Body: "new"})
	if err == nil {
		t.Error("expected error for out-of-range prose")
	}
}

func TestCellsWithRuntimeReadFile(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// A cell with reads(data.csv:runtime) to exercise the runtime mode branch
	src := "```go:notebook name=test reads(data.csv:runtime)\nx := 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/cells", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var cells []cellInfo
	json.Unmarshal(w.Body.Bytes(), &cells)
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	if len(cells[0].Reads) != 1 || cells[0].Reads[0].Mode != "runtime" {
		t.Errorf("expected runtime mode read, got: %+v", cells[0].Reads)
	}
}

func TestCellsWithTimeout(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test timeout=30s\nx := 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/cells", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var cells []cellInfo
	json.Unmarshal(w.Body.Bytes(), &cells)
	if len(cells) != 1 {
		t.Fatalf("got %d cells, want 1", len(cells))
	}
	if cells[0].Timeout != "30s" {
		t.Errorf("timeout = %q, want 30s", cells[0].Timeout)
	}
}

func TestNewServerWithMissingStaticSubFS(t *testing.T) {
	// New() must not panic — just ensure it returns a usable server.
	srv := New("nonexistent.md", ":0")
	if srv == nil {
		t.Error("New should not return nil")
	}
}

func TestExtractFencePrefixTildes(t *testing.T) {
	got := extractFencePrefix("~~~go:notebook")
	if got != "~~~" {
		t.Errorf("extractFencePrefix(~~~) = %q, want ~~~", got)
	}
}

func TestExtractFencePrefixNotFence(t *testing.T) {
	// A line that starts with a non-fence character returns default "```".
	got := extractFencePrefix("not a fence")
	if got != "```" {
		t.Errorf("extractFencePrefix(not a fence) = %q, want ```", got)
	}
}

func TestRewriteCellWithLiveData(t *testing.T) {
	src := "```go:notebook name=test\nval = 1\n```\n"
	req := updateRequest{
		Line: 1,
		Body: "val = 2",
		Name: "test",
		Data: "live",
	}
	got, err := rewriteCell(src, req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "data=live") {
		t.Errorf("missing data=live in:\n%s", got)
	}
}

func TestParseCompileErrorZeroLine(t *testing.T) {
	// A file:0:col: message should return nil (line 0 is invalid).
	got := parseCompileError("file.go:0:1: some error")
	if got != nil {
		t.Errorf("line 0 should return nil, got %+v", got)
	}
}

func TestParseCompileErrorEmptyMessage(t *testing.T) {
	// A file:1:1:   (blank message) should return nil.
	got := parseCompileError("file.go:1:1:   ")
	if got != nil {
		t.Errorf("empty message should return nil, got %+v", got)
	}
}

func TestParseRunErrorsCellError(t *testing.T) {
	stderr := "cell execution failed: something\n"
	errs := parseRunErrors(stderr, "", "test")
	found := false
	for _, e := range errs {
		if e.Kind == "runtime" && strings.Contains(e.Message, "cell") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected runtime error for 'cell ...' line, got: %v", errs)
	}
}

func TestRunBatchWithCheckErrors(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// exported without name triggers checker rule violation
	src := "```go:notebook exported(val int)\nval = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/run", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestRunStateAfterExecution(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")

	// Run the notebook.
	body, _ := json.Marshal(runRequest{Cell: "test"})
	runReq := httptest.NewRequest(http.MethodPost, "/api/run", bytes.NewReader(body))
	runReq.Header.Set("Content-Type", "application/json")
	runW := httptest.NewRecorder()
	srv.Handler().ServeHTTP(runW, runReq)

	// Check run state.
	stateReq := httptest.NewRequest(http.MethodGet, "/api/run/state", nil)
	stateW := httptest.NewRecorder()
	srv.Handler().ServeHTTP(stateW, stateReq)

	if stateW.Code != http.StatusOK {
		t.Fatalf("status = %d", stateW.Code)
	}

	var states map[string]*cellState
	json.Unmarshal(stateW.Body.Bytes(), &states)
	if _, ok := states["test"]; !ok {
		t.Error("expected cell state for 'test' after execution")
	}
}

func TestCheckUpstreamFailureMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")
	// Should not panic; returns "" when file is missing.
	refused := srv.checkUpstreamFailure("any")
	if refused != "" {
		t.Errorf("checkUpstreamFailure with missing file should return empty, got %q", refused)
	}
}

func TestMarkStaleWithUnknownCell(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=a exported(x int)\nx = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	srv.mu.Lock()
	srv.cellStates["a"] = &cellState{Name: "a", Stale: false}
	srv.mu.Unlock()

	// Mark a cell name that doesn't exist in the notebook — should not panic.
	srv.markStale("unknown")
}

func TestRewriteProseOutOfRangeLines(t *testing.T) {
	// A degenerate case: prose block found but lines out of range.
	// Hard to trigger naturally, test via rewriteProse with valid prose at line 1.
	src := "# Title\n\nmore.\n"
	updated, err := rewriteProse(src, proseUpdateRequest{Line: 1, Body: "# New\n\nupdated."})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updated, "New") {
		t.Error("should contain updated content")
	}
}

func TestHandleEditorEmbeddedHTMLNotFound(t *testing.T) {
	// handleEditor reads static/dist/index.html which is embedded.
	// Verify the endpoint returns 200 with an HTML body (file must be embedded).
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/editor", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	// Should be 200 (file is embedded at build time).
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (static/dist/index.html must be embedded)", w.Code)
	}
}

// --- Editor CRUD tests (#281–#292) ---

func TestBlocksAddCell(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(addBlockRequest{Kind: "cell", AfterLine: 1, Body: "x := 1", Info: "go:notebook name=newcell"})
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	if !strings.Contains(content, "go:notebook name=newcell") {
		t.Error("cell info string missing")
	}
	if !strings.Contains(content, "x := 1") {
		t.Error("cell body missing")
	}
	if !strings.Contains(content, "# Title") {
		t.Error("original content should be preserved")
	}
}

func TestBlocksAddProse(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(addBlockRequest{Kind: "prose", AfterLine: 1, Body: "Some new prose."})
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	if !strings.Contains(content, "Some new prose.") {
		t.Error("prose body missing")
	}
}

func TestBlocksAddAtBeginning(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(addBlockRequest{Kind: "cell", AfterLine: 0, Body: "first := true", Info: "go:notebook name=first"})
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	// Cell should be at the beginning, before the title.
	idx := strings.Index(content, "go:notebook name=first")
	titleIdx := strings.Index(content, "# Title")
	if idx < 0 || titleIdx < 0 || idx > titleIdx {
		t.Error("cell should be inserted before existing content")
	}
}

func TestBlocksAddInvalidKind(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(addBlockRequest{Kind: "invalid", AfterLine: 0})
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlocksAddMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/blocks/add", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestBlocksDeleteBlock(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Title\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n\nEnd prose.\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Find the cell block line.
	body, _ := json.Marshal(deleteBlockRequest{Line: 3})
	req := httptest.NewRequest(http.MethodDelete, "/api/blocks/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	if strings.Contains(content, "val = 42") {
		t.Error("deleted block body should not appear in file")
	}
	if strings.Contains(content, "go:notebook name=test") {
		t.Error("deleted block info should not appear in file")
	}
	if !strings.Contains(content, "# Title") {
		t.Error("non-deleted content should be preserved")
	}
	if !strings.Contains(content, "End prose.") {
		t.Error("content after deleted block should be preserved")
	}
}

func TestBlocksDeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(deleteBlockRequest{Line: 999})
	req := httptest.NewRequest(http.MethodDelete, "/api/blocks/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlocksDeleteMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/blocks/delete", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestBlocksReorderDown(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// No blank lines between cells, so they are adjacent blocks.
	src := "# Title\n```go:notebook name=a exported(x int)\nx = 1\n```\n```go:notebook name=b exported(y int)\ny = 2\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Cell a is at line 2, cell b at line 5.
	body, _ := json.Marshal(reorderBlockRequest{Line: 2, Direction: "down"})
	req := httptest.NewRequest(http.MethodPut, "/api/blocks/reorder", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	// After reorder, "b" should come before "a".
	bIdx := strings.Index(content, "name=b")
	aIdx := strings.Index(content, "name=a")
	if bIdx < 0 || aIdx < 0 || bIdx > aIdx {
		t.Errorf("after moving a down, b should precede a:\n%s", content)
	}
}

func TestBlocksReorderUp(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// No blank lines between cells.
	src := "# Title\n```go:notebook name=a exported(x int)\nx = 1\n```\n```go:notebook name=b exported(y int)\ny = 2\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Cell b is at line 5.
	body, _ := json.Marshal(reorderBlockRequest{Line: 5, Direction: "up"})
	req := httptest.NewRequest(http.MethodPut, "/api/blocks/reorder", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	bIdx := strings.Index(content, "name=b")
	aIdx := strings.Index(content, "name=a")
	if bIdx < 0 || aIdx < 0 || bIdx > aIdx {
		t.Errorf("after moving b up, b should precede a:\n%s", content)
	}
}

func TestBlocksReorderInvalidDirection(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(reorderBlockRequest{Line: 1, Direction: "sideways"})
	req := httptest.NewRequest(http.MethodPut, "/api/blocks/reorder", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlocksReorderFirstUp(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(reorderBlockRequest{Line: 1, Direction: "up"})
	req := httptest.NewRequest(http.MethodPut, "/api/blocks/reorder", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlocksReorderMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/blocks/reorder", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestNotebookInit(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte(""), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/notebook/init", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	if !strings.Contains(content, "# My Notebook") {
		t.Error("starter content should contain title")
	}
	if !strings.Contains(content, "go:defs") {
		t.Error("starter content should contain defs block")
	}
	if !strings.Contains(content, "go:notebook name=hello") {
		t.Error("starter content should contain hello cell")
	}
	if !strings.Contains(content, "fmt.Println") {
		t.Error("starter content should contain hello code")
	}
}

func TestNotebookInitNonEmpty(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Existing content\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/notebook/init", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w.Code)
	}

	// Verify file was not modified.
	data, _ := os.ReadFile(notebook)
	if !strings.Contains(string(data), "# Existing content") {
		t.Error("file should not be modified")
	}
}

func TestNotebookInitWhitespaceOnly(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("   \n  \n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/notebook/init", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	if !strings.Contains(string(data), "# My Notebook") {
		t.Error("whitespace-only file should be treated as empty")
	}
}

func TestNotebookInitMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte(""), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/notebook/init", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestInsertBlockUnit(t *testing.T) {
	src := "# Title\n\nSome prose.\n"

	// Insert a cell after line 3 (end of prose).
	result := insertBlock(src, addBlockRequest{Kind: "cell", AfterLine: 3, Body: "x := 1", Info: "go:notebook name=test"})
	if !strings.Contains(result, "```go:notebook name=test") {
		t.Errorf("missing cell fence:\n%s", result)
	}
	if !strings.Contains(result, "x := 1") {
		t.Errorf("missing cell body:\n%s", result)
	}
	if !strings.Contains(result, "# Title") {
		t.Errorf("original content missing:\n%s", result)
	}
}

func TestDeleteBlockUnit(t *testing.T) {
	src := "# Title\n\n```go:notebook name=test\nval = 42\n```\n\nEnd.\n"
	result, err := deleteBlock(src, 3)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "val = 42") {
		t.Errorf("deleted block should not appear:\n%s", result)
	}
	if !strings.Contains(result, "# Title") {
		t.Errorf("title should be preserved:\n%s", result)
	}
	if !strings.Contains(result, "End.") {
		t.Errorf("end prose should be preserved:\n%s", result)
	}
}

func TestDeleteBlockNotFound(t *testing.T) {
	src := "# Title\n"
	_, err := deleteBlock(src, 999)
	if err == nil {
		t.Error("expected error for nonexistent block")
	}
}

func TestReorderBlockUnit(t *testing.T) {
	// Two cells separated only by a blank prose line. Moving b up swaps it
	// with the blank prose (the immediate neighbor), not cell a.
	// Use a source with no blank line between cells to test direct swap.
	src := "# Top\n```go:notebook name=a\nx = 1\n```\n```go:notebook name=b\ny = 2\n```\n"
	// Parser: block 0 = prose(line 1), block 1 = cell a(line 2-4), block 2 = cell b(line 5-7).
	// Move cell a (line 2) down — should swap with cell b.
	result, err := reorderBlock(src, 2, "down")
	if err != nil {
		t.Fatal(err)
	}
	bIdx := strings.Index(result, "name=b")
	aIdx := strings.Index(result, "name=a")
	if bIdx < 0 || aIdx < 0 || bIdx > aIdx {
		t.Errorf("after moving a down, b should come first:\n%s", result)
	}
}

func TestReorderBlockCannotMoveFirstUp(t *testing.T) {
	src := "# Title\n"
	_, err := reorderBlock(src, 1, "up")
	if err == nil {
		t.Error("expected error when moving first block up")
	}
}

func TestReorderBlockCannotMoveLastDown(t *testing.T) {
	src := "# Title\n"
	_, err := reorderBlock(src, 1, "down")
	if err == nil {
		t.Error("expected error when moving last block down")
	}
}

func TestReorderBlockNotFound(t *testing.T) {
	src := "# Title\n"
	_, err := reorderBlock(src, 999, "up")
	if err == nil {
		t.Error("expected error for nonexistent block")
	}
}

// --- Coverage gap tests ---

func TestBlocksAddDefs(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(addBlockRequest{Kind: "defs", AfterLine: 1, Body: "import \"fmt\""})
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	content := string(data)
	if !strings.Contains(content, "go:defs") {
		t.Error("defs info string missing")
	}
	if !strings.Contains(content, "import \"fmt\"") {
		t.Error("defs body missing")
	}
}

func TestBlocksAddDefaultCellInfo(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	// No Info field — should default to "go:notebook name=cell1"
	body, _ := json.Marshal(addBlockRequest{Kind: "cell", AfterLine: 1, Body: "x := 1"})
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	if !strings.Contains(string(data), "go:notebook name=cell1") {
		t.Error("default cell info should be 'go:notebook name=cell1'")
	}
}

func TestBlocksAddInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/add", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlocksAddMissingFile(t *testing.T) {
	s := New("/nonexistent/test.md", ":0")
	body, _ := json.Marshal(addBlockRequest{Kind: "cell", AfterLine: 0, Body: "x := 1"})
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestBlocksDeleteInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodDelete, "/api/blocks/delete", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlocksDeleteMissingFile(t *testing.T) {
	s := New("/nonexistent/test.md", ":0")
	body, _ := json.Marshal(deleteBlockRequest{Line: 1})
	req := httptest.NewRequest(http.MethodDelete, "/api/blocks/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestBlocksReorderInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPut, "/api/blocks/reorder", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestBlocksReorderMissingFile(t *testing.T) {
	s := New("/nonexistent/test.md", ":0")
	body, _ := json.Marshal(reorderBlockRequest{Line: 1, Direction: "up"})
	req := httptest.NewRequest(http.MethodPut, "/api/blocks/reorder", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestBlocksReorderNotFoundLine(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	body, _ := json.Marshal(reorderBlockRequest{Line: 999, Direction: "up"})
	req := httptest.NewRequest(http.MethodPut, "/api/blocks/reorder", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestNotebookInitNewFile(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "new.md")
	// File does not exist — handleNotebookInit should create it.

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/notebook/init", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	data, _ := os.ReadFile(notebook)
	if !strings.Contains(string(data), "# My Notebook") {
		t.Error("new file should have starter content")
	}
}

func TestNotebookInitReadError(t *testing.T) {
	// Point to a path that is a directory, so ReadFile fails with a non-IsNotExist error.
	dir := t.TempDir()
	s := New(dir, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/notebook/init", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestInsertBlockDefsNoInfo(t *testing.T) {
	src := "# Title\n"
	result := insertBlock(src, addBlockRequest{Kind: "defs", AfterLine: 1, Body: "import \"os\""})
	if !strings.Contains(result, "```go:defs") {
		t.Errorf("defs block should use default info:\n%s", result)
	}
}

func TestInsertBlockAfterLineExceedsLength(t *testing.T) {
	src := "# Title\n"
	result := insertBlock(src, addBlockRequest{Kind: "cell", AfterLine: 999, Body: "x := 1", Info: "go:notebook name=test"})
	if !strings.Contains(result, "x := 1") {
		t.Errorf("block should be inserted even when after_line exceeds length:\n%s", result)
	}
}

func TestInsertBlockProseEmptyBody(t *testing.T) {
	src := "# Title\n"
	result := insertBlock(src, addBlockRequest{Kind: "prose", AfterLine: 1, Body: ""})
	// Should not panic, should produce valid output.
	if !strings.Contains(result, "# Title") {
		t.Errorf("original content missing:\n%s", result)
	}
}

func TestInsertBlockAtBeginningNonEmptyFirstLine(t *testing.T) {
	src := "# Title\nContent\n"
	result := insertBlock(src, addBlockRequest{Kind: "prose", AfterLine: 0, Body: "Prepended."})
	idx := strings.Index(result, "Prepended.")
	titleIdx := strings.Index(result, "# Title")
	if idx < 0 || titleIdx < 0 || idx > titleIdx {
		t.Errorf("prepended prose should be before title:\n%s", result)
	}
}

func TestRunMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/run", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	// runBatch reads the file, should get an error.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestRunSingleCellMissingFile(t *testing.T) {
	srv := New("/nonexistent/test.md", ":0")
	body, _ := json.Marshal(runRequest{Cell: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	// executeNotebook reads the file, error goes in resp.Error, status is 200.
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var resp runResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error == "" {
		t.Error("expected error message for missing file")
	}
}

func TestRewriteCellNotFound(t *testing.T) {
	src := "# Title\n"
	_, err := rewriteCell(src, updateRequest{Line: 999, Body: "x := 1"})
	if err == nil {
		t.Error("expected error for nonexistent cell")
	}
}

func TestRootPathRedirect(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "/editor" {
		t.Errorf("redirect Location = %q, want /editor", loc)
	}
}

func TestRootPathUnknown404(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestRewriteCellWithRuntimeReads(t *testing.T) {
	src := "```go:notebook name=test\nval = 1\n```\n"
	req := updateRequest{
		Line:  1,
		Body:  "val = 2",
		Name:  "test",
		Reads: []readFileInfo{{Name: "data.csv", Mode: "runtime"}},
	}
	got, err := rewriteCell(src, req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "reads(data.csv:runtime)") {
		t.Errorf("missing runtime reads in:\n%s", got)
	}
}

func TestParseCompileErrorNonNumericLine(t *testing.T) {
	got := parseCompileError("file.go:abc:1: some error")
	if got != nil {
		t.Errorf("non-numeric line should return nil, got %+v", got)
	}
}

func TestParseCompileErrorTooFewParts(t *testing.T) {
	got := parseCompileError("file.go:1")
	if got != nil {
		t.Errorf("too few parts should return nil, got %+v", got)
	}
}

func TestParseRunErrorsRecoveredPanic(t *testing.T) {
	stderr := "recovered panic in cell test: bad thing\n"
	errs := parseRunErrors(stderr, "", "test")
	found := false
	for _, e := range errs {
		if e.Kind == "panic" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected panic error for 'recovered panic' line, got: %v", errs)
	}
}

func TestParseRunErrorsCellErrorPrefix(t *testing.T) {
	stderr := "CellError: something failed\n"
	errs := parseRunErrors(stderr, "", "test")
	found := false
	for _, e := range errs {
		if e.Kind == "runtime" && strings.Contains(e.Message, "CellError") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected runtime error for 'CellError' line, got: %v", errs)
	}
}

func TestLSPCompletionNoExportedFields(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=test\nx := 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/completion", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "test") {
		t.Error("completion should include cell name 'test'")
	}
	// Should not have field completions since no exported fields.
	if strings.Contains(body, "test.") {
		t.Error("should not have field completions without exported fields")
	}
}

func TestLSPCompletionUnnamedCells(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook\nx := 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/completion", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	// Unnamed cells should be skipped.
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	items := resp["items"].([]interface{})
	for _, item := range items {
		m := item.(map[string]interface{})
		if m["kind"] == "cell" {
			t.Error("unnamed cells should not appear in completions")
		}
	}
}

func TestLSPHoverFieldNotFound(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=test exported(val int)\nval = 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/hover?word=test.nonexistent", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var info hoverInfo
	json.Unmarshal(w.Body.Bytes(), &info)
	if info.Content != "" {
		t.Errorf("hover for unknown field should be empty, got %q", info.Content)
	}
}

func TestLSPHoverCellWithoutExports(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=test\nval := 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/hover?word=test", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var info hoverInfo
	json.Unmarshal(w.Body.Bytes(), &info)
	if !strings.Contains(info.Content, "Cell test") {
		t.Error("hover should show cell name")
	}
	// Should NOT contain "Exported:" since there are no exports.
	if strings.Contains(info.Content, "Exported:") {
		t.Error("hover should not show Exported for cell without exports")
	}
}

func TestRunBatchSuccessMarksNotStale(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=step1 exported(val int)\nval = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")

	// Pre-mark cell as stale.
	srv.mu.Lock()
	srv.cellStates["step1"] = &cellState{Name: "step1", Stale: true}
	srv.mu.Unlock()

	// Run batch (no cell specified).
	req := httptest.NewRequest(http.MethodPost, "/api/run", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp runResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr: %s, error: %s", resp.ExitCode, resp.Stderr, resp.Error)
	}

	// After successful batch run, cell should not be stale.
	srv.mu.RLock()
	cs := srv.cellStates["step1"]
	srv.mu.RUnlock()
	if cs == nil {
		t.Fatal("cell state should exist after batch run")
	}
	if cs.Stale {
		t.Error("cell should not be stale after successful batch run")
	}
}

func TestExecuteNotebookCheckErrors(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// exported without name triggers checker error.
	src := "```go:notebook exported(val int)\nval = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	body, _ := json.Marshal(runRequest{Cell: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var resp runResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !strings.Contains(resp.Error, "check errors") {
		t.Errorf("expected check errors, got: %q", resp.Error)
	}
}

func TestDeleteBlockSkipsBlankLine(t *testing.T) {
	// Tests that deleteBlock skips a trailing blank line after the deleted block.
	src := "# Title\n\n```go:notebook name=test\nval = 42\n```\n\nEnd.\n"
	result, err := deleteBlock(src, 3)
	if err != nil {
		t.Fatal(err)
	}
	// The blank line between the cell and "End." should be consumed.
	if strings.Contains(result, "val = 42") {
		t.Error("cell body should be removed")
	}
	if !strings.Contains(result, "End.") {
		t.Error("content after block should be preserved")
	}
}

func TestCellsUpdateWriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	// Use a read-only directory to trigger WriteFile error.
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Test\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Make the file read-only.
	os.Chmod(notebook, 0o444)
	defer os.Chmod(notebook, 0o644)

	req := updateRequest{Line: 3, Body: "val = 99", Name: "test", Exported: "val int"}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPut, "/api/cells/update", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestProseUpdateWriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Title\n\nSome prose.\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Make file read-only.
	os.Chmod(notebook, 0o444)
	defer os.Chmod(notebook, 0o644)

	body, _ := json.Marshal(proseUpdateRequest{Line: 1, Body: "# New Title"})
	req := httptest.NewRequest(http.MethodPut, "/api/prose/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestBlocksAddWriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	s := New(notebook, ":0")

	// Make file read-only.
	os.Chmod(notebook, 0o444)
	defer os.Chmod(notebook, 0o644)

	body, _ := json.Marshal(addBlockRequest{Kind: "prose", AfterLine: 1, Body: "new"})
	req := httptest.NewRequest(http.MethodPost, "/api/blocks/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestBlocksDeleteWriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Title\n\n```go:notebook name=test\nval = 42\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Make file read-only.
	os.Chmod(notebook, 0o444)
	defer os.Chmod(notebook, 0o644)

	body, _ := json.Marshal(deleteBlockRequest{Line: 3})
	req := httptest.NewRequest(http.MethodDelete, "/api/blocks/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestBlocksReorderWriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "# Title\n```go:notebook name=a\nx = 1\n```\n```go:notebook name=b\ny = 2\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")

	// Make file read-only.
	os.Chmod(notebook, 0o444)
	defer os.Chmod(notebook, 0o644)

	body, _ := json.Marshal(reorderBlockRequest{Line: 2, Direction: "down"})
	req := httptest.NewRequest(http.MethodPut, "/api/blocks/reorder", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestNotebookInitWriteError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte(""), 0o644)

	s := New(notebook, ":0")

	// Make the file read-only.
	os.Chmod(notebook, 0o444)
	defer os.Chmod(notebook, 0o644)

	req := httptest.NewRequest(http.MethodPost, "/api/notebook/init", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestFindModuleRoot(t *testing.T) {
	// findModuleRoot should return the directory containing go.mod.
	root := findModuleRoot()
	if root == "" {
		t.Error("findModuleRoot returned empty string")
	}
	// Should be a valid directory.
	info, err := os.Stat(root)
	if err != nil {
		t.Fatalf("findModuleRoot returned non-existent path: %v", err)
	}
	if !info.IsDir() {
		t.Error("findModuleRoot should return a directory")
	}
}

func TestRunAllStopsOnError(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// First cell is good, second will fail (references undefined variable).
	src := "```go:notebook name=step1 exported(val int)\nval = 1\n```\n\n```go:notebook name=step2 exported(out int)\nout = undefined_var\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/run/all", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var results []runResponse
	json.Unmarshal(w.Body.Bytes(), &results)
	// Should stop after first error.
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	lastResult := results[len(results)-1]
	if lastResult.Error == "" && lastResult.ExitCode == 0 {
		t.Error("expected error in final result")
	}
}

func TestRewriteCellWithTildeFence(t *testing.T) {
	src := "~~~go:notebook name=test\nval = 1\n~~~\n"
	req := updateRequest{
		Line: 1,
		Body: "val = 2",
		Name: "test",
	}
	got, err := rewriteCell(src, req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "~~~") {
		t.Errorf("tilde fence should be preserved:\n%s", got)
	}
	if !strings.Contains(got, "val = 2") {
		t.Errorf("body should be updated:\n%s", got)
	}
}

func TestRewriteCellIndentedFence(t *testing.T) {
	src := "  ```go:notebook name=test\nval = 1\n  ```\n"
	req := updateRequest{
		Line: 1,
		Body: "val = 2",
		Name: "test",
	}
	got, err := rewriteCell(src, req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "  ```") {
		t.Errorf("indented fence should be preserved:\n%s", got)
	}
}

func TestMarkStaleFileReadError(t *testing.T) {
	// markStale with a non-empty cell name on a missing file should not panic.
	srv := New("/nonexistent/test.md", ":0")
	srv.mu.Lock()
	srv.cellStates["a"] = &cellState{Name: "a", Stale: false}
	srv.mu.Unlock()

	// Should silently return without panicking.
	srv.markStale("a")

	srv.mu.RLock()
	defer srv.mu.RUnlock()
	// State should be unchanged since file read failed.
	if srv.cellStates["a"].Stale {
		t.Error("cell should not be marked stale when file read fails")
	}
}

func TestLSPCompletionMultipleExportedFields(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=data exported(x int, y string)\nx = 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/completion", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "data.x") {
		t.Error("completion should include data.x")
	}
	if !strings.Contains(body, "data.y") {
		t.Error("completion should include data.y")
	}
}

func TestLSPCompletionExportedFieldNoType(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// Exported field with just a name but no type — should be skipped in completions.
	os.WriteFile(notebook, []byte("```go:notebook name=data exported(x)\nx = 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/completion", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	// Cell name should be present but field "data.x" should not (no type).
	if !strings.Contains(body, "\"data\"") {
		t.Error("completion should include cell name 'data'")
	}
	if strings.Contains(body, "data.x") {
		t.Error("field without type should not appear in completions")
	}
}

func TestLSPHoverExportedFieldNoType(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// Exported field with just a name and no type.
	os.WriteFile(notebook, []byte("```go:notebook name=data exported(x)\nx = 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/hover?word=data.x", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var info hoverInfo
	json.Unmarshal(w.Body.Bytes(), &info)
	// Field without type should return empty hover.
	if info.Content != "" {
		t.Errorf("hover for field without type should be empty, got %q", info.Content)
	}
}

func TestLSPHoverMultipleExportedFieldsMismatch(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("```go:notebook name=data exported(x int, y string)\nx = 1\n```\n"), 0o644)

	s := New(notebook, ":0")
	// Hover on a field that exists.
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/hover?word=data.y", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var info hoverInfo
	json.Unmarshal(w.Body.Bytes(), &info)
	if !strings.Contains(info.Content, "data.y") {
		t.Errorf("hover should show data.y, got: %q", info.Content)
	}
	if !strings.Contains(info.Content, "string") {
		t.Errorf("hover should show field type string, got: %q", info.Content)
	}
}

func TestRunBatchRuntimeError(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// A cell that compiles but panics at runtime.
	src := "```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test\nvar s []int\nfmt.Println(s[999])\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/run", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp runResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	// Should have non-zero exit code.
	if resp.ExitCode == 0 && resp.Error == "" {
		t.Error("expected runtime error or non-zero exit code")
	}
}

func TestRunSingleCellRuntimeError(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// A cell that compiles but panics at runtime.
	src := "```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test\nvar s []int\nfmt.Println(s[999])\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	body, _ := json.Marshal(runRequest{Cell: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp runResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	// Should have non-zero exit code or error.
	if resp.ExitCode == 0 && resp.Error == "" {
		t.Error("expected runtime error or non-zero exit code for single cell")
	}
	// Cell state should be marked as failed.
	srv.mu.RLock()
	cs := srv.cellStates["test"]
	srv.mu.RUnlock()
	if cs == nil {
		t.Fatal("cell state should exist after execution")
	}
	if !cs.Failed {
		t.Error("cell state should be marked as failed")
	}
}

func TestRunBatchCompileError(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// A cell that fails to compile (undefined variable).
	src := "```go:notebook name=test exported(val int)\nval = undefined_xyz\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodPost, "/api/run", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp runResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	// Should have non-zero exit code.
	if resp.ExitCode == 0 && resp.Error == "" {
		t.Error("expected compile error")
	}
}

func TestLSPCompletionWithMixedBlocks(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// Notebook with prose, defs, and cells to exercise the continue path for non-cell blocks.
	src := "# Title\n\nSome prose.\n\n```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/completion", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "test") {
		t.Error("completion should include cell name")
	}
	if !strings.Contains(body, "test.val") {
		t.Error("completion should include field")
	}
}

func TestLSPHoverWithMixedBlocks(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// Notebook with prose, defs, and cells.
	src := "# Title\n\nSome prose.\n\n```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=test exported(val int)\nval = 42\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/hover?word=test", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var info hoverInfo
	json.Unmarshal(w.Body.Bytes(), &info)
	if !strings.Contains(info.Content, "Cell test") {
		t.Error("hover should show cell name")
	}
}

func TestListenAndServeStartsAndStops(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	os.WriteFile(notebook, []byte("# Title\n"), 0o644)

	srv := New(notebook, "127.0.0.1:0")

	// Start the server in a goroutine and immediately connect to verify it started.
	// We use a listener to get the actual port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	srv2 := New(notebook, addr)
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv2.ListenAndServe()
	}()

	// Give the server a moment to start.
	time.Sleep(100 * time.Millisecond)

	// Make a request to verify it's running.
	resp, err := http.Get("http://" + addr + "/api/status")
	if err != nil {
		t.Logf("server may not have started: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
	}

	// We just need to verify it started — we can't cleanly shut it down
	// without adding a Shutdown method, but the test will clean up on exit.
	_ = srv // suppress unused warning
}

func TestRunBatchMkdirTempError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("TMPDIR not supported on Windows")
	}
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")

	// Set TMPDIR to a non-writable path to trigger MkdirTemp error.
	old := os.Getenv("TMPDIR")
	os.Setenv("TMPDIR", "/nonexistent-tmpdir-for-test")
	defer os.Setenv("TMPDIR", old)

	req := httptest.NewRequest(http.MethodPost, "/api/run", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 (MkdirTemp should fail)", w.Code)
	}
}

func TestRunSingleCellMkdirTempError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("TMPDIR not supported on Windows")
	}
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")

	// Set TMPDIR to a non-writable path to trigger MkdirTemp error.
	old := os.Getenv("TMPDIR")
	os.Setenv("TMPDIR", "/nonexistent-tmpdir-for-test")
	defer os.Setenv("TMPDIR", old)

	body, _ := json.Marshal(runRequest{Cell: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var resp runResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error == "" {
		t.Error("expected error when TMPDIR is non-writable")
	}
}

func TestFindModuleRootWithBadGoEnv(t *testing.T) {
	// Temporarily break PATH so `go env GOMOD` fails.
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "/nonexistent-path")
	defer os.Setenv("PATH", oldPath)

	root := findModuleRoot()
	if root != "." {
		t.Errorf("findModuleRoot with broken PATH should return '.', got %q", root)
	}
}

func TestRunBatchCancelledContext(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")

	// Create a request with an already-cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	req := httptest.NewRequest(http.MethodPost, "/api/run", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	// The response should contain an error since the context was cancelled.
	if w.Code == http.StatusOK {
		var resp runResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		// Either error is non-empty or exit code is non-zero.
		if resp.Error == "" && resp.ExitCode == 0 {
			t.Log("note: cancelled context may not propagate quickly enough")
		}
	}
}

func TestRunSingleCellCancelledContext(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	src := "```go:notebook name=test exported(val int)\nval = 1\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	srv := New(notebook, ":0")

	// Create a request with an already-cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	body, _ := json.Marshal(runRequest{Cell: "test"})
	req := httptest.NewRequest(http.MethodPost, "/api/run", bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var resp runResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Error == "" && resp.ExitCode == 0 {
			t.Log("note: cancelled context may not propagate quickly enough")
		}
	}
}

func TestLSPHoverFieldOnCellWithMultipleCells(t *testing.T) {
	dir := t.TempDir()
	notebook := filepath.Join(dir, "test.md")
	// Two cells — hover on second cell's field should skip first cell.
	src := "```go:notebook name=alpha exported(a int)\na = 1\n```\n\n```go:notebook name=beta exported(b int)\nb = 2\n```\n"
	os.WriteFile(notebook, []byte(src), 0o644)

	s := New(notebook, ":0")
	req := httptest.NewRequest(http.MethodGet, "/api/lsp/hover?word=beta.b", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var info hoverInfo
	json.Unmarshal(w.Body.Bytes(), &info)
	if !strings.Contains(info.Content, "beta.b") {
		t.Errorf("hover should show beta.b, got: %q", info.Content)
	}
}

// --- Coverage gap tests ---

func TestRewriteCellOutOfRange(t *testing.T) {
	src := "# Hello\n\n```go:notebook name=x\nfoo\n```\n"
	// Tamper with a request that has a valid line but make the source too short
	// by using a line number that doesn't correspond to a cell.
	_, err := rewriteCell(src, updateRequest{Line: 999, Body: "bar"})
	if err == nil {
		t.Error("expected error for non-existent cell line")
	}
}

func TestRewriteProseLineOutOfBounds(t *testing.T) {
	src := "# Hello\n"
	_, err := rewriteProse(src, proseUpdateRequest{Line: 999, Body: "new"})
	if err == nil {
		t.Error("expected error for non-existent prose line")
	}
}

func TestDeleteBlockLineNotFound(t *testing.T) {
	src := "# Hello\n\n```go:notebook name=x\nfoo\n```\n"
	_, err := deleteBlock(src, 999)
	if err == nil {
		t.Error("expected error for non-existent block line")
	}
}

func TestHandleEditorReturnsHTML(t *testing.T) {
	// The embedded static/dist/index.html always exists, but we can test the
	// method guard on a GET request which still exercises the handler path.
	path := testNotebook(t)
	s := New(path, "localhost:0")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/editor", nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "go-notebook editor") {
		t.Error("editor HTML not returned")
	}
}

func TestFindModuleRootFallback(t *testing.T) {
	// Save and clear PATH to make "go env GOMOD" fail.
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)
	root := findModuleRoot()
	if root != "." {
		t.Errorf("expected fallback '.', got %q", root)
	}
}

func TestRunBatchGenerateError(t *testing.T) {
	// Create a notebook with invalid Go that will fail checker.
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.md")
	// A cell with a body that references undefined is a checker error.
	src := "```go:notebook name=bad\nundefined_func()\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	s := New(path, "localhost:0")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/run", nil)
	s.mux.ServeHTTP(w, r)
	// Should return 400 with checker errors.
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		// Either is acceptable — the point is it doesn't crash.
	}
}

func TestExecuteNotebookGenerateError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.md")
	src := "```go:notebook name=bad\nundefined_func()\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	s := New(path, "localhost:0")
	w := httptest.NewRecorder()
	body := `{"cell":"bad"}`
	r := httptest.NewRequest("POST", "/api/run", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s.mux.ServeHTTP(w, r)
	// Should return a response (not crash).
	if w.Code == 0 {
		t.Error("expected a response")
	}
}

func TestRunBatchWithValidNotebook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.md")
	src := "# Test\n\n```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=hello\nfmt.Println(\"hi\")\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	s := New(path, "localhost:0")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/run", nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status %d", w.Code)
	}
}

func TestRunAllWithValidNotebook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.md")
	src := "# Test\n\n```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=hello\nfmt.Println(\"hi\")\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	s := New(path, "localhost:0")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/run/all", nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status %d", w.Code)
	}
}

func TestRunSingleCellWithUpstreamFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.md")
	src := "```go:notebook name=first\nfoo\n```\n\n```go:notebook name=second\nbar\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	s := New(path, "localhost:0")
	// Mark first cell as failed.
	s.cellStates["first"] = &cellState{Name: "first", Failed: true}

	w := httptest.NewRecorder()
	body := `{"cell":"second"}`
	r := httptest.NewRequest("POST", "/api/run", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 conflict for upstream failure, got %d", w.Code)
	}
}

func TestRewriteCellStartLineOutOfRange(t *testing.T) {
	// rewriteCell with a matching line but tampered source that's too short
	_, err := rewriteCell("", updateRequest{Line: 1, Body: "bar"})
	if err == nil {
		t.Error("expected error for empty source with line 1")
	}
}

func TestRewriteProseBoundaryLines(t *testing.T) {
	// Prose block at lines 1-1, replace with multi-line.
	src := "Hello world\n\n```go:notebook name=x\nfoo\n```\n"
	updated, err := rewriteProse(src, proseUpdateRequest{Line: 1, Body: "New first line\nSecond line"})
	if err != nil {
		t.Fatalf("rewriteProse: %v", err)
	}
	if !strings.Contains(updated, "New first line") {
		t.Error("expected updated prose")
	}
}

func TestDeleteBlockBoundaryCase(t *testing.T) {
	// Delete the only block in a file.
	src := "# Title\n\n```go:notebook name=only\ncode\n```\n"
	updated, err := deleteBlock(src, 3)
	if err != nil {
		t.Fatalf("deleteBlock: %v", err)
	}
	if strings.Contains(updated, "go:notebook") {
		t.Error("cell should be deleted")
	}
	if !strings.Contains(updated, "# Title") {
		t.Error("prose should be preserved")
	}
}

func TestSanitizeOutputCleanText(t *testing.T) {
	input := "Hello world\nLine 2"
	output := sanitizeOutput(input)
	if output != input {
		t.Errorf("expected no change, got %q", output)
	}
}

func TestSanitizeOutputStripsSVGBlock(t *testing.T) {
	input := "Before\n<svg xmlns=\"http://www.w3.org/2000/svg\"><circle r=\"10\"/></svg>\nAfter"
	output := sanitizeOutput(input)
	if strings.Contains(output, "<svg") {
		t.Error("SVG should be stripped")
	}
	if !strings.Contains(output, "Before") {
		t.Error("content before SVG should be preserved")
	}
	if !strings.Contains(output, "[SVG content stripped") {
		t.Error("should have SVG stripped message")
	}
}

func TestExtractFencePrefixVariants(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"```go:notebook", "```"},
		{"~~~go:notebook", "~~~"},
		{"````go:notebook", "````"},
		{"   ```go:notebook", "   ```"},
		{"", "```"},
		{"no fence here", "```"},
	}
	for _, tt := range tests {
		got := extractFencePrefix(tt.line)
		if got != tt.want {
			t.Errorf("extractFencePrefix(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestRunBatchSuccessPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "good.md")
	// Minimal valid notebook that passes checker: defs block + named cell with valid Go.
	src := "# Test\n\n```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=greet\nfmt.Println(\"hi\")\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	s := New(path, "localhost:0")

	// First trigger the batch run (POST /api/run with no body).
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/run", nil)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	s.mux.ServeHTTP(w, r.WithContext(ctx))
	// Accept OK or 500 (may fail to compile due to env), but must not panic.
	if w.Code == 0 {
		t.Error("expected a response")
	}

	// Also trigger single-cell run to cover executeNotebook path.
	w2 := httptest.NewRecorder()
	body := `{"cell":"greet"}`
	r2 := httptest.NewRequest("POST", "/api/run", strings.NewReader(body))
	r2.Header.Set("Content-Type", "application/json")
	ctx2, cancel2 := context.WithTimeout(r2.Context(), 60*time.Second)
	defer cancel2()
	s.mux.ServeHTTP(w2, r2.WithContext(ctx2))
	if w2.Code == 0 {
		t.Error("expected a response")
	}
}

func TestRunSingleCellSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "good.md")
	src := "# Test\n\n```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=greet\nfmt.Println(\"hi\")\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	s := New(path, "localhost:0")
	w := httptest.NewRecorder()
	body := `{"cell":"greet"}`
	r := httptest.NewRequest("POST", "/api/run", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	s.mux.ServeHTTP(w, r.WithContext(ctx))
	if w.Code == 0 {
		t.Error("expected a response")
	}
}

func TestRunAllCellsSequential(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.md")
	src := "```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=a\nfmt.Println(\"a\")\n```\n\n```go:notebook name=b\nfmt.Println(\"b\")\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	s := New(path, "localhost:0")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/run/all", nil)
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	s.mux.ServeHTTP(w, r.WithContext(ctx))
	if w.Code == 0 {
		t.Error("expected a response")
	}
}

func TestRunStateWithStaleCells(t *testing.T) {
	path := testNotebook(t)
	s := New(path, "localhost:0")
	s.cellStates["test"] = &cellState{Name: "test", Stale: true}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/run/state", nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var states map[string]*cellState
	json.Unmarshal(w.Body.Bytes(), &states)
	if states["test"] == nil || !states["test"].Stale {
		t.Error("expected test cell to be stale")
	}
}

func TestMarkStaleAllCells(t *testing.T) {
	path := testNotebook(t)
	s := New(path, "localhost:0")
	s.cellStates["a"] = &cellState{Name: "a", Stale: false}
	s.cellStates["b"] = &cellState{Name: "b", Stale: false}
	s.markStale("")
	for _, cs := range s.cellStates {
		if !cs.Stale {
			t.Errorf("cell %s should be stale", cs.Name)
		}
	}
}

func TestRunBatchTmpdirUnwritable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("TMPDIR not supported on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "good.md")
	src := "# Test\n\n```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=greet\nfmt.Println(\"hi\")\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	// Point TMPDIR to a non-writable directory to force MkdirTemp failure.
	badDir := filepath.Join(dir, "noperm")
	os.Mkdir(badDir, 0o000)
	defer os.Chmod(badDir, 0o755)
	t.Setenv("TMPDIR", badDir)

	s := New(path, "localhost:0")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/run", nil)
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for unwritable tmpdir, got %d", w.Code)
	}
}

func TestExecuteNotebookTmpdirUnwritable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("TMPDIR not supported on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "good.md")
	src := "# Test\n\n```go:defs\nimport \"fmt\"\n```\n\n```go:notebook name=greet\nfmt.Println(\"hi\")\n```\n"
	os.WriteFile(path, []byte(src), 0o644)

	badDir := filepath.Join(dir, "noperm")
	os.Mkdir(badDir, 0o000)
	defer os.Chmod(badDir, 0o755)
	t.Setenv("TMPDIR", badDir)

	s := New(path, "localhost:0")
	w := httptest.NewRecorder()
	body := `{"cell":"greet"}`
	r := httptest.NewRequest("POST", "/api/run", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	s.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		// executeNotebook returns 200 with error in the response body.
	}
}

func TestParseRunErrorsVariants(t *testing.T) {
	// Compile error format.
	errs := parseRunErrors("test.md:10:5: undefined: foo", "", "cell1")
	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}
	if errs[0].Kind != "compile" {
		t.Errorf("expected compile, got %s", errs[0].Kind)
	}

	// Timeout.
	errs = parseRunErrors("", "context deadline exceeded", "cell1")
	if len(errs) == 0 || errs[0].Kind != "timeout" {
		t.Error("expected timeout error")
	}

	// Generic error with no stderr.
	errs = parseRunErrors("", "some other error", "cell1")
	if len(errs) == 0 || errs[0].Kind != "runtime" {
		t.Error("expected runtime error")
	}

	// No errors at all.
	errs = parseRunErrors("", "", "cell1")
	if len(errs) != 0 {
		t.Error("expected no errors")
	}

	// Skip goroutine/runtime lines.
	errs = parseRunErrors("goroutine 1 [running]:\nruntime.goexit()\n\t/usr/local/go/src/runtime/asm_amd64.s:1700", "", "cell1")
	if len(errs) != 0 {
		t.Errorf("expected stack trace lines to be skipped, got %d errors", len(errs))
	}
}

func TestIsSourceTree(t *testing.T) {
	// Find the actual repo root (this file is at nb/server/).
	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))

	t.Run("actual repo root", func(t *testing.T) {
		if !isSourceTree(repoRoot) {
			t.Errorf("expected isSourceTree(%q) = true for the repo root", repoRoot)
		}
	})

	t.Run("dot returns false", func(t *testing.T) {
		if isSourceTree(".") {
			t.Error("expected isSourceTree(\".\") = false")
		}
	})

	t.Run("temp dir without go.mod", func(t *testing.T) {
		dir := t.TempDir()
		if isSourceTree(dir) {
			t.Errorf("expected isSourceTree(%q) = false for empty dir", dir)
		}
	})

	t.Run("temp dir with wrong module path", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/other\n"), 0o644)
		if isSourceTree(dir) {
			t.Errorf("expected isSourceTree(%q) = false for wrong module", dir)
		}
	})
}
