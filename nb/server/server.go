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
	s.mux.HandleFunc("/api/cells/update", s.handleCellsUpdate)
	s.mux.HandleFunc("/api/run", s.handleRun)
	s.mux.HandleFunc("/api/editor", s.handleEditor)

	// Serve static files at /static/.
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

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

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
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
