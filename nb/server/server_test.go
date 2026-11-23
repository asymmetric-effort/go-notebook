// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	r := httptest.NewRequest(http.MethodGet, "/api/editor", nil)
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

	r := httptest.NewRequest(http.MethodPost, "/api/editor", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want 405", w.Code)
	}
}

func TestStaticFileServing(t *testing.T) {
	path := testNotebook(t)
	srv := New(path, ":0")

	r := httptest.NewRequest(http.MethodGet, "/static/editor.js", nil)
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
