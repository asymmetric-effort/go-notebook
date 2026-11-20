// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleStatus(t *testing.T) {
	s := New(":0", "test.md")
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	s.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["notebook"] != "test.md" {
		t.Errorf("notebook = %v", resp["notebook"])
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %v", resp["status"])
	}
}

func TestHandleRunGetRejected(t *testing.T) {
	s := New(":0", "test.md")
	req := httptest.NewRequest("GET", "/api/run", nil)
	w := httptest.NewRecorder()
	s.handleRun(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestHandleRunPost(t *testing.T) {
	s := New(":0", "test.md")
	req := httptest.NewRequest("POST", "/api/run", nil)
	w := httptest.NewRecorder()
	s.handleRun(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestNewServer(t *testing.T) {
	s := New(":0", "test.md")
	if s.Addr() != ":0" {
		t.Errorf("Addr() = %q", s.Addr())
	}
}

func TestListenAndServeCancel(t *testing.T) {
	s := New("127.0.0.1:0", "test.md")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.
	err := s.ListenAndServe(ctx)
	// Should return quickly (server closed).
	if err != nil && err != http.ErrServerClosed {
		t.Logf("ListenAndServe error (expected): %v", err)
	}
}
