// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package server implements the gobook serve HTTP server. It provides a
// notebook editing and execution environment accessible from a browser.
package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
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

//go:embed static
var staticFiles embed.FS

// cellState tracks per-cell execution state.
type cellState struct {
	Name     string `json:"name"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
	Stale    bool   `json:"stale"`
	Failed   bool   `json:"failed"`
}

// Server is the gobook serve HTTP server.
type Server struct {
	addr         string
	notebookPath string
	mux          *http.ServeMux
	kernel       *kernel.Kernel
	mu           sync.RWMutex
	lastParse    core.ParseResult
	cellStates   map[string]*cellState // Per-cell execution state by name.
}

// New creates a new server for the given notebook.
func New(notebookPath, addr string) *Server {
	s := &Server{
		addr:         addr,
		notebookPath: notebookPath,
		mux:          http.NewServeMux(),
		kernel:       kernel.New(),
		cellStates:   make(map[string]*cellState),
	}
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/cells", s.handleCells)
	s.mux.HandleFunc("/api/cells/update", s.handleCellsUpdate)
	s.mux.HandleFunc("/api/blocks", s.handleBlocks)
	s.mux.HandleFunc("/api/prose/update", s.handleProseUpdate)
	s.mux.HandleFunc("/api/run", s.handleRun)
	s.mux.HandleFunc("/api/run/all", s.handleRunAll)
	s.mux.HandleFunc("/api/run/state", s.handleRunState)
	s.mux.HandleFunc("/api/editor", s.handleEditor)

	// Serve static files at /static/.
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

	return s
}

// cspPolicy is the Content-Security-Policy for editor pages.
// It blocks all remote loads, inline scripts, and object/embed/frame elements.
const cspPolicy = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; object-src 'none'; frame-src 'none'; base-uri 'self'"

// Handler returns the HTTP handler with CSP middleware for testing.
func (s *Server) Handler() http.Handler {
	return cspMiddleware(s.mux)
}

// cspMiddleware wraps a handler to set the Content-Security-Policy header
// on every response.
func cspMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", cspPolicy)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           cspMiddleware(s.mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	fmt.Fprintf(os.Stderr, "gobook serve: %s on %s\n", s.notebookPath, s.addr)
	return srv.ListenAndServe()
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

// readFileInfo is a serialisable ReadFile for the JSON API.
type readFileInfo struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
}

// cellInfo describes a cell for the /api/cells endpoint.
type cellInfo struct {
	Name     string         `json:"name"`
	Line     int            `json:"line"`
	Exported string         `json:"exported"`
	Body     string         `json:"body"`
	Reads    []readFileInfo `json:"reads"`
	Writes   []string       `json:"writes"`
	Secrets  []string       `json:"secrets"`
	Data     string         `json:"data"`
	Timeout  string         `json:"timeout"`
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
	cells := make([]cellInfo, 0)
	for _, b := range result.Blocks {
		if b.Kind == core.BlockCell {
			attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)

			reads := make([]readFileInfo, len(attrs.Reads))
			for i, rf := range attrs.Reads {
				mode := "embed"
				if rf.Mode == core.FileModeRuntime {
					mode = "runtime"
				}
				reads[i] = readFileInfo{Name: rf.Name, Mode: mode}
			}
			writes := attrs.Writes
			if writes == nil {
				writes = []string{}
			}
			secrets := attrs.Secrets
			if secrets == nil {
				secrets = []string{}
			}

			timeout := ""
			if attrs.Timeout != 0 {
				timeout = attrs.Timeout.String()
			}

			cells = append(cells, cellInfo{
				Name:     attrs.Name,
				Line:     b.StartPos.Line,
				Exported: attrs.Exported,
				Body:     b.Body,
				Reads:    reads,
				Writes:   writes,
				Secrets:  secrets,
				Data:     attrs.Data.String(),
				Timeout:  timeout,
			})
		}
	}

	writeJSON(w, http.StatusOK, cells)
}

// updateRequest is the JSON body for PUT /api/cells/update.
type updateRequest struct {
	Line     int            `json:"line"`
	Body     string         `json:"body"`
	Name     string         `json:"name"`
	Exported string         `json:"exported"`
	Reads    []readFileInfo `json:"reads"`
	Writes   []string       `json:"writes"`
	Secrets  []string       `json:"secrets"`
	Data     string         `json:"data"`
	Timeout  string         `json:"timeout"`
}

func (s *Server) handleCellsUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}

	src, err := os.ReadFile(s.notebookPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	updated, err := rewriteCell(string(src), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := os.WriteFile(s.notebookPath, []byte(updated), 0o644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Mark this cell and all downstream cells as stale.
	s.markStale(req.Name)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// markStale marks a cell and all cells that come after it as stale.
func (s *Server) markStale(cellName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cellName == "" {
		// Mark all cells stale.
		for _, cs := range s.cellStates {
			cs.Stale = true
		}
		return
	}

	// Parse to find ordering. Mark the named cell and everything after it stale.
	src, err := os.ReadFile(s.notebookPath)
	if err != nil {
		return
	}
	result := core.Parse(string(src))
	found := false
	for _, b := range result.Blocks {
		if b.Kind != core.BlockCell {
			continue
		}
		attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)
		if attrs.Name == cellName {
			found = true
		}
		if found && attrs.Name != "" {
			if cs, ok := s.cellStates[attrs.Name]; ok {
				cs.Stale = true
			}
		}
	}
}

// rewriteCell replaces a cell identified by its opening fence line number with
// a new info string and body, leaving all other content intact.
func rewriteCell(src string, req updateRequest) (string, error) {
	result := core.Parse(src)

	var target *core.Block
	for i := range result.Blocks {
		b := &result.Blocks[i]
		if b.Kind == core.BlockCell && b.StartPos.Line == req.Line {
			target = b
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("no cell found at line %d", req.Line)
	}

	// Build updated CellAttrs from the request.
	attrs := core.CellAttrs{
		Name:     req.Name,
		Exported: req.Exported,
		Timeout:  0,
	}
	for _, rf := range req.Reads {
		mode := core.FileModeEmbed
		if rf.Mode == "runtime" {
			mode = core.FileModeRuntime
		}
		attrs.Reads = append(attrs.Reads, core.ReadFile{Name: rf.Name, Mode: mode})
	}
	attrs.Writes = req.Writes
	attrs.Secrets = req.Secrets
	switch req.Data {
	case "embedded":
		attrs.Data = core.DataEmbedded
	case "live":
		attrs.Data = core.DataLive
	default:
		attrs.Data = core.DataNone
	}
	if req.Timeout != "" {
		dur, err := time.ParseDuration(req.Timeout)
		if err != nil {
			return "", fmt.Errorf("invalid timeout %q: %w", req.Timeout, err)
		}
		attrs.Timeout = dur
	}

	newInfo := core.FormatInfoString(attrs)

	// Reconstruct the file by replacing the target cell's lines.
	// We identify the opening fence line (StartPos.Line, 1-based) and closing
	// fence line (EndPos.Line, 1-based) and replace everything between them.
	lines := strings.Split(src, "\n")

	// Determine the fence character and length from the original opening line.
	openIdx := target.StartPos.Line - 1 // 0-based
	closeIdx := target.EndPos.Line - 1

	if openIdx < 0 || openIdx >= len(lines) {
		return "", fmt.Errorf("cell start line %d out of range", target.StartPos.Line)
	}

	// Extract fence prefix from original opening line.
	origOpen := lines[openIdx]
	fencePrefix := extractFencePrefix(origOpen)

	// Build replacement lines.
	var replacement []string
	replacement = append(replacement, fencePrefix+newInfo)
	// Body lines (body may contain embedded newlines).
	bodyLines := strings.Split(req.Body, "\n")
	replacement = append(replacement, bodyLines...)
	replacement = append(replacement, fencePrefix)

	// Replace lines [openIdx, closeIdx] inclusive.
	var out []string
	out = append(out, lines[:openIdx]...)
	out = append(out, replacement...)
	if closeIdx+1 < len(lines) {
		out = append(out, lines[closeIdx+1:]...)
	}

	return strings.Join(out, "\n"), nil
}

// extractFencePrefix returns the fence characters (backticks/tildes with any
// leading indent) from an opening fence line, stripping the info string.
func extractFencePrefix(line string) string {
	indent := 0
	for indent < len(line) && indent < 3 && line[indent] == ' ' {
		indent++
	}
	if indent >= len(line) {
		return "```"
	}
	ch := line[indent]
	if ch != '`' && ch != '~' {
		return "```"
	}
	count := 0
	for indent+count < len(line) && line[indent+count] == ch {
		count++
	}
	return line[:indent+count]
}

func (s *Server) handleEditor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := staticFiles.ReadFile("static/editor.html")
	if err != nil {
		http.Error(w, "editor not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

// blockInfo describes a block for the /api/blocks endpoint.
type blockInfo struct {
	Kind     string `json:"kind"`               // "prose", "cell", "defs", "output"
	Line     int    `json:"line"`               // Start line (1-based).
	EndLine  int    `json:"end_line"`           // End line (1-based).
	Body     string `json:"body"`               // Block content.
	Info     string `json:"info,omitempty"`     // Info string for fenced blocks.
	Name     string `json:"name,omitempty"`     // Cell name (cells only).
	Exported string `json:"exported,omitempty"` // Exported fields (cells only).
}

func (s *Server) handleBlocks(w http.ResponseWriter, r *http.Request) {
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
	var blocks []blockInfo
	for _, b := range result.Blocks {
		bi := blockInfo{
			Kind:    b.Kind.String(),
			Line:    b.StartPos.Line,
			EndLine: b.EndPos.Line,
			Body:    b.Body,
			Info:    b.Info,
		}
		if b.Kind == core.BlockCell {
			attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)
			bi.Name = attrs.Name
			bi.Exported = attrs.Exported
		}
		blocks = append(blocks, bi)
	}

	writeJSON(w, http.StatusOK, blocks)
}

// proseUpdateRequest is the JSON body for PUT /api/prose/update.
type proseUpdateRequest struct {
	Line int    `json:"line"` // Start line of the prose block (1-based).
	Body string `json:"body"` // New prose content.
}

func (s *Server) handleProseUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req proseUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
		return
	}

	src, err := os.ReadFile(s.notebookPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	updated, err := rewriteProse(string(src), req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := os.WriteFile(s.notebookPath, []byte(updated), 0o644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// rewriteProse replaces a prose block's content in the source, preserving
// all other blocks (cells, defs, output) untouched.
func rewriteProse(src string, req proseUpdateRequest) (string, error) {
	result := core.Parse(src)

	var target *core.Block
	for i := range result.Blocks {
		b := &result.Blocks[i]
		if b.Kind == core.BlockProse && b.StartPos.Line == req.Line {
			target = b
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("no prose block found at line %d", req.Line)
	}

	lines := strings.Split(src, "\n")
	startIdx := target.StartPos.Line - 1 // 0-based
	endIdx := target.EndPos.Line - 1

	if startIdx < 0 || endIdx >= len(lines) {
		return "", fmt.Errorf("prose block lines %d-%d out of range", target.StartPos.Line, target.EndPos.Line)
	}

	// Replace lines [startIdx, endIdx] inclusive with new body.
	newLines := strings.Split(req.Body, "\n")

	var out []string
	out = append(out, lines[:startIdx]...)
	out = append(out, newLines...)
	if endIdx+1 < len(lines) {
		out = append(out, lines[endIdx+1:]...)
	}

	return strings.Join(out, "\n"), nil
}

// runRequest is the JSON body for POST /api/run.
type runRequest struct {
	Cell string `json:"cell,omitempty"` // Cell name to run (empty = run all).
}

// cellError is a parsed error with cell and line context.
type cellError struct {
	Message string `json:"message"`        // Human-readable error message.
	Cell    string `json:"cell,omitempty"` // Cell name where the error occurred.
	File    string `json:"file,omitempty"` // Source file (notebook.md).
	Line    int    `json:"line,omitempty"` // Markdown line number (1-based).
	Kind    string `json:"kind,omitempty"` // "compile", "runtime", "panic", "timeout"
}

// runResponse is the JSON response for /api/run.
type runResponse struct {
	Cell     string      `json:"cell,omitempty"`
	Stdout   string      `json:"stdout"`
	Stderr   string      `json:"stderr"`
	ExitCode int         `json:"exit_code"`
	Error    string      `json:"error,omitempty"`
	Refused  string      `json:"refused,omitempty"` // Name of failed upstream if refused.
	Errors   []cellError `json:"errors,omitempty"`  // Parsed errors with cell/line context.
}

// handleRun runs a single cell or all cells. Accepts optional JSON body
// with {"cell":"name"}. Without a cell name, runs all cells as a batch
// and returns a single combined response.
func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req runRequest
	if r.ContentLength > 0 {
		json.NewDecoder(r.Body).Decode(&req)
	}

	if req.Cell != "" {
		s.runSingleCell(w, r, req.Cell)
	} else {
		s.runBatch(w, r)
	}
}

// runBatch runs the entire notebook as a single go run invocation
// and returns one combined response.
func (s *Server) runBatch(w http.ResponseWriter, r *http.Request) {
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

	// Mark all cells as not-stale and not-failed on successful batch run.
	if cellResult.ExitCode == 0 && runErr == nil {
		s.mu.Lock()
		for _, b := range result.Blocks {
			if b.Kind == core.BlockCell {
				attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)
				if attrs.Name != "" {
					s.cellStates[attrs.Name] = &cellState{Name: attrs.Name, Stale: false, Failed: false}
				}
			}
		}
		s.mu.Unlock()
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleRunAll runs all cells sequentially, stopping on first error.
func (s *Server) handleRunAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.runAllCells(w, r)
}

// handleRunState returns the current execution state of all cells.
func (s *Server) handleRunState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	states := make(map[string]*cellState, len(s.cellStates))
	for k, v := range s.cellStates {
		cp := *v
		states[k] = &cp
	}
	s.mu.RUnlock()

	writeJSON(w, http.StatusOK, states)
}

// runSingleCell executes one named cell, refusing if an upstream has failed.
func (s *Server) runSingleCell(w http.ResponseWriter, r *http.Request, cellName string) {
	// Check for failed upstream.
	if refused := s.checkUpstreamFailure(cellName); refused != "" {
		writeJSON(w, http.StatusConflict, runResponse{
			Cell:    cellName,
			Refused: refused,
			Error:   fmt.Sprintf("refused: upstream cell %q has failed", refused),
		})
		return
	}

	resp := s.executeNotebook(r, cellName)
	writeJSON(w, http.StatusOK, resp)
}

// runAllCells executes all cells sequentially, stopping on first error.
func (s *Server) runAllCells(w http.ResponseWriter, r *http.Request) {
	src, err := os.ReadFile(s.notebookPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	result := core.Parse(string(src))

	// Collect named cells in order.
	var cellNames []string
	for _, b := range result.Blocks {
		if b.Kind == core.BlockCell {
			attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)
			if attrs.Name != "" {
				cellNames = append(cellNames, attrs.Name)
			}
		}
	}

	var results []runResponse
	for _, name := range cellNames {
		resp := s.executeNotebook(r, name)
		results = append(results, resp)
		if resp.ExitCode != 0 || resp.Error != "" {
			// Stop on first error.
			break
		}
	}

	writeJSON(w, http.StatusOK, results)
}

// checkUpstreamFailure returns the name of the first failed upstream cell,
// or empty string if all upstreams are ok.
func (s *Server) checkUpstreamFailure(cellName string) string {
	src, err := os.ReadFile(s.notebookPath)
	if err != nil {
		return ""
	}
	result := core.Parse(string(src))

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, b := range result.Blocks {
		if b.Kind != core.BlockCell {
			continue
		}
		attrs, _ := core.ParseInfoString(b.Info, b.StartPos.Line)
		if attrs.Name == cellName {
			break // Reached the target cell — no failed upstreams.
		}
		if attrs.Name != "" {
			if cs, ok := s.cellStates[attrs.Name]; ok && cs.Failed {
				return attrs.Name
			}
		}
	}
	return ""
}

// executeNotebook generates, compiles and runs the notebook, returning results
// for the specified cell (or all cells if cellName is empty).
func (s *Server) executeNotebook(r *http.Request, cellName string) runResponse {
	src, err := os.ReadFile(s.notebookPath)
	if err != nil {
		return runResponse{Cell: cellName, Error: err.Error()}
	}

	result := core.Parse(string(src))
	errs := core.Check(result)
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		return runResponse{Cell: cellName, Error: "check errors: " + strings.Join(msgs, "; ")}
	}

	gen, err := core.Generate(result, "main")
	if err != nil {
		return runResponse{Cell: cellName, Error: err.Error()}
	}

	tmpDir, err := os.MkdirTemp("", "gobook-serve-*")
	if err != nil {
		return runResponse{Cell: cellName, Error: err.Error()}
	}
	defer os.RemoveAll(tmpDir)

	pkgDir := filepath.Join(tmpDir, "pkg")
	if err := core.WritePackage(gen, pkgDir, s.notebookPath, "gobook-serve"); err != nil {
		return runResponse{Cell: cellName, Error: err.Error()}
	}

	moduleRoot := findModuleRoot()
	goModPath := filepath.Join(pkgDir, "go.mod")
	goModData, _ := os.ReadFile(goModPath)
	goModStr := string(goModData) + "\nreplace github.com/asymmetric-effort/go-notebook => " + moduleRoot + "\n"
	os.WriteFile(goModPath, []byte(goModStr), 0o644)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	cellResult, runErr := s.kernel.RunCell(ctx, cellName, 5*time.Minute,
		func(ctx context.Context) (kernel.CellResult, error) {
			return kernel.GoRun(ctx, pkgDir)
		})

	resp := runResponse{
		Cell:     cellName,
		Stdout:   cellResult.Stdout,
		Stderr:   cellResult.Stderr,
		ExitCode: cellResult.ExitCode,
	}
	if runErr != nil {
		resp.Error = runErr.Error()
	}

	// Parse errors from stderr and the error message.
	resp.Errors = parseRunErrors(cellResult.Stderr, resp.Error, cellName)

	// Sanitize output: strip raw SVG content from stdout.
	resp.Stdout = sanitizeOutput(resp.Stdout)

	// Update cell state.
	s.mu.Lock()
	s.cellStates[cellName] = &cellState{
		Name:     cellName,
		Stdout:   cellResult.Stdout,
		Stderr:   cellResult.Stderr,
		ExitCode: cellResult.ExitCode,
		Error:    resp.Error,
		Stale:    false,
		Failed:   cellResult.ExitCode != 0 || runErr != nil,
	}
	s.mu.Unlock()

	return resp
}

// parseRunErrors extracts structured errors from stderr output and error
// messages. It recognizes Go compile errors with //line directives,
// panics caught by deferred recovery, and timeout errors.
// Stack traces are stripped to show only the actionable message.
func parseRunErrors(stderr, errMsg, cellName string) []cellError {
	var errs []cellError

	// Check for timeout.
	if strings.Contains(errMsg, "timed out") || strings.Contains(errMsg, "context deadline") {
		errs = append(errs, cellError{
			Message: "Cell timed out. Increase the timeout or check for infinite loops.",
			Cell:    cellName,
			Kind:    "timeout",
		})
		return errs
	}

	// Parse compile errors from stderr (format: "file:line:col: message").
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip stack traces (goroutine lines, runtime lines).
		if strings.HasPrefix(line, "goroutine ") || strings.HasPrefix(line, "\t") ||
			strings.Contains(line, "runtime.") || strings.Contains(line, "runtime/") {
			continue
		}
		// Parse "notebook.md:LINE:COL: message" format from //line directives.
		if ce := parseCompileError(line); ce != nil {
			errs = append(errs, *ce)
			continue
		}
		// Parse gobook cell error format.
		if strings.HasPrefix(line, "cell ") || strings.Contains(line, "CellError") {
			errs = append(errs, cellError{
				Message: line,
				Cell:    cellName,
				Kind:    "runtime",
			})
			continue
		}
		// Parse panic messages from deferred recovery.
		if strings.Contains(line, "panic:") || strings.Contains(line, "recovered panic") {
			msg := line
			// Strip "panic: " prefix for cleaner display.
			if idx := strings.Index(msg, "panic: "); idx >= 0 {
				msg = msg[idx+7:]
			}
			errs = append(errs, cellError{
				Message: "Panic: " + msg,
				Cell:    cellName,
				Kind:    "panic",
			})
			continue
		}
	}

	// If we have an error message but no parsed errors, add a generic one.
	if errMsg != "" && len(errs) == 0 {
		errs = append(errs, cellError{
			Message: errMsg,
			Cell:    cellName,
			Kind:    "runtime",
		})
	}

	return errs
}

// parseCompileError parses a Go compiler error line with //line directives.
// Format: "notebook.md:LINE:COL: message" or "./main.go:LINE:COL: message"
func parseCompileError(line string) *cellError {
	// Look for "file:line:col: message" pattern.
	parts := strings.SplitN(line, ":", 4)
	if len(parts) < 4 {
		return nil
	}
	file := parts[0]
	lineNum := 0
	for _, ch := range parts[1] {
		if ch >= '0' && ch <= '9' {
			lineNum = lineNum*10 + int(ch-'0')
		} else {
			return nil
		}
	}
	if lineNum == 0 {
		return nil
	}
	msg := strings.TrimSpace(parts[3])
	if msg == "" {
		return nil
	}

	return &cellError{
		Message: msg,
		File:    file,
		Line:    lineNum,
		Kind:    "compile",
	}
}

// sanitizeOutput strips raw SVG content from cell output to prevent
// untrusted SVG from reaching the browser. SVG can contain embedded
// scripts and event handlers. Only tool-generated chart SVG (which is
// rendered via the chart package) is safe; raw SVG from cell stdout is not.
func sanitizeOutput(output string) string {
	// Simple heuristic: if output contains an SVG tag, strip the SVG block.
	lower := strings.ToLower(output)
	if !strings.Contains(lower, "<svg") {
		return output
	}

	var result strings.Builder
	lines := strings.Split(output, "\n")
	inSVG := false
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if strings.Contains(lower, "<svg") {
			inSVG = true
			result.WriteString("[SVG content stripped for security]\n")
			continue
		}
		if inSVG {
			if strings.Contains(lower, "</svg>") {
				inSVG = false
			}
			continue
		}
		result.WriteString(line)
		result.WriteString("\n")
	}
	return strings.TrimRight(result.String(), "\n")
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
