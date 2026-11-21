// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package server implements the gobook serve HTTP server. It provides a
// notebook editing and execution environment accessible from a browser.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/asymmetric-effort/go-notebook/nb/core"
	"github.com/asymmetric-effort/go-notebook/nb/kernel"
)

// Server is the gobook serve HTTP server.
type Server struct {
	addr         string
	notebookPath string
	mux          *http.ServeMux
	kernel       *kernel.Kernel
	mu           sync.RWMutex
	lastParse    core.ParseResult
}

// New creates a new server for the given notebook.
func New(notebookPath, addr string) *Server {
	s := &Server{
		addr:         addr,
		notebookPath: notebookPath,
		mux:          http.NewServeMux(),
		kernel:       kernel.New(),
	}
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/cells", s.handleCells)
	s.mux.HandleFunc("/api/run", s.handleRun)
	return s
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	fmt.Fprintf(os.Stderr, "gobook serve: %s on %s\n", s.notebookPath, s.addr)
	return srv.ListenAndServe()
}

// Handler returns the HTTP handler for testing.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// statusResponse is the JSON response for /api/status.
type statusResponse struct {
	Notebook string `json:"notebook"`
	Cells    int    `json:"cells"`
	OK       bool   `json:"ok"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	src, err := os.ReadFile(s.notebookPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	result := core.Parse(string(src))
	cellCount := 0
	for _, b := range result.Blocks {
		if b.Kind == core.BlockCell {
			cellCount++
		}
	}

	s.mu.Lock()
	s.lastParse = result
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, statusResponse{
		Notebook: filepath.Base(s.notebookPath),
		Cells:    cellCount,
		OK:       true,
	})
}

// cellInfo describes a cell for the /api/cells endpoint.
type cellInfo struct {
	Name     string `json:"name,omitempty"`
	Line     int    `json:"line"`
	Exported string `json:"exported,omitempty"`
}

func (s *Server) handleCells(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	src, err := os.ReadFile(s.notebookPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	result := core.Parse(string(src))
	var cells []cellInfo
	for _, b := range result.Blocks {
		if b.Kind == core.BlockCell {
			attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)
			cells = append(cells, cellInfo{
				Name:     attrs.Name,
				Line:     b.StartPos.Line,
				Exported: attrs.Exported,
			})
		}
	}

	writeJSON(w, http.StatusOK, cells)
}

// runResponse is the JSON response for /api/run.
type runResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	src, err := os.ReadFile(s.notebookPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errors": msgs,
		})
		return
	}

	gen, err := core.Generate(result, "main")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	tmpDir, err := os.MkdirTemp("", "gobook-serve-*")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer os.RemoveAll(tmpDir)

	pkgDir := filepath.Join(tmpDir, "pkg")
	if err := core.WritePackage(gen, pkgDir, s.notebookPath, "gobook-serve"); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Add replace directive.
	moduleRoot := findModuleRoot()
	goModPath := filepath.Join(pkgDir, "go.mod")
	goModData, _ := os.ReadFile(goModPath)
	goModStr := string(goModData) + "\nreplace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(goModPath, []byte(goModStr), 0o644)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	cellResult, runErr := s.kernel.RunCell(ctx, "all", 5*time.Minute,
		func(ctx context.Context) (kernel.CellResult, error) {
			return kernel.GoRun(ctx, pkgDir)
		})

	resp := runResponse{
		Stdout:   cellResult.Stdout,
		Stderr:   cellResult.Stderr,
		ExitCode: cellResult.ExitCode,
	}
	if runErr != nil {
		resp.Error = runErr.Error()
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func findModuleRoot() string {
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		return "."
	}
	modFile := strings.TrimSpace(string(out))
	if modFile == "" || modFile == os.DevNull {
		return "."
	}
	return filepath.Dir(modFile)
}
