// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
