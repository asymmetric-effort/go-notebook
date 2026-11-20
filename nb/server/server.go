// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

// Package server implements gobook serve: the notebook server that runs
// cells through the kernel and streams output to the browser.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/asymmetric-effort/go-notebook/nb/kernel"
)

// Server is the notebook server for interactive editing.
type Server struct {
	addr     string
	kernel   *kernel.Kernel
	mu       sync.Mutex
	notebook string
	mux      *http.ServeMux
}

// New creates a new notebook server.
func New(addr, notebook string) *Server {
	s := &Server{
		addr:     addr,
		kernel:   kernel.New(),
		notebook: notebook,
		mux:      http.NewServeMux(),
	}
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/run", s.handleRun)
	return s
}

// ListenAndServe starts the server.
func (s *Server) ListenAndServe(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	srv := &http.Server{Handler: s.mux}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	fmt.Printf("gobook serve: %s at http://%s\n", s.notebook, listener.Addr())
	return srv.Serve(listener)
}

// Addr returns the server address (for testing).
func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"notebook":  s.notebook,
		"cacheSize": s.kernel.CacheSize(),
		"status":    "ok",
	})
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	// Placeholder: would parse cell name from request and run it.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "not implemented",
	})
}
