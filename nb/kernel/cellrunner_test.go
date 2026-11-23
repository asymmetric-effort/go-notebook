// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package kernel

import (
	"net"
	"testing"
)

func TestNewCellRunner(t *testing.T) {
	cr := NewCellRunner()
	if cr == nil {
		t.Fatal("NewCellRunner returned nil")
	}
	if len(cr.cache) != 0 {
		t.Errorf("cache len = %d, want 0", len(cr.cache))
	}
	if len(cr.displays) != 0 {
		t.Errorf("displays len = %d, want 0", len(cr.displays))
	}
}

func TestCellRunnerCache(t *testing.T) {
	cr := NewCellRunner()

	// Miss.
	_, ok := cr.GetCacheEntry("load")
	if ok {
		t.Error("expected cache miss")
	}

	// Set.
	cr.SetCacheEntry("load", []byte("data-load"))
	data, ok := cr.GetCacheEntry("load")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(data) != "data-load" {
		t.Errorf("data = %q, want %q", data, "data-load")
	}

	// Overwrite.
	cr.SetCacheEntry("load", []byte("updated"))
	data, ok = cr.GetCacheEntry("load")
	if !ok {
		t.Fatal("expected cache hit after overwrite")
	}
	if string(data) != "updated" {
		t.Errorf("data = %q, want %q", data, "updated")
	}
}

func TestCellRunnerGetDisplays(t *testing.T) {
	cr := NewCellRunner()
	displays := cr.GetDisplays()
	if len(displays) != 0 {
		t.Errorf("displays = %d, want 0", len(displays))
	}
}

func TestHandleConnectionFetchUpstream(t *testing.T) {
	cr := NewCellRunner()
	cr.SetCacheEntry("load", []byte("cached-load-data"))

	server, client := net.Pipe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- cr.handleConnection(server, "transform")
	}()

	// Cell requests upstream "load".
	if err := WriteMessage(client, ProtocolMessage{
		Type: MsgFetchUpstream,
		Cell: "load",
	}); err != nil {
		t.Fatal(err)
	}

	// Read the response — should be upstream_data.
	resp, err := ReadMessage(client)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Type != MsgUpstreamData {
		t.Errorf("resp.Type = %q, want %q", resp.Type, MsgUpstreamData)
	}
	if resp.Cell != "load" {
		t.Errorf("resp.Cell = %q, want %q", resp.Cell, "load")
	}
	if string(resp.Data) != "cached-load-data" {
		t.Errorf("resp.Data = %q, want %q", resp.Data, "cached-load-data")
	}

	// Send result to end the session.
	if err := WriteMessage(client, ProtocolMessage{
		Type: MsgResult,
		Cell: "transform",
		Data: []byte("transform-result"),
	}); err != nil {
		t.Fatal(err)
	}

	client.Close()
	if err := <-errCh; err != nil {
		t.Fatalf("handleConnection: %v", err)
	}

	// Verify the result was cached.
	data, ok := cr.GetCacheEntry("transform")
	if !ok {
		t.Fatal("transform result not cached")
	}
	if string(data) != "transform-result" {
		t.Errorf("cached data = %q, want %q", data, "transform-result")
	}
}

func TestHandleConnectionFetchUpstreamMiss(t *testing.T) {
	cr := NewCellRunner()

	server, client := net.Pipe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- cr.handleConnection(server, "transform")
	}()

	// Request an upstream cell that's not in cache.
	if err := WriteMessage(client, ProtocolMessage{
		Type: MsgFetchUpstream,
		Cell: "missing",
	}); err != nil {
		t.Fatal(err)
	}

	// Should get an error response.
	resp, err := ReadMessage(client)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Type != MsgError {
		t.Errorf("resp.Type = %q, want %q", resp.Type, MsgError)
	}
	if resp.Err == "" {
		t.Error("expected non-empty error message")
	}

	// Send result to end the session.
	if err := WriteMessage(client, ProtocolMessage{
		Type: MsgResult,
		Cell: "transform",
		Data: []byte("done"),
	}); err != nil {
		t.Fatal(err)
	}

	client.Close()
	if err := <-errCh; err != nil {
		t.Fatalf("handleConnection: %v", err)
	}
}

func TestHandleConnectionDisplay(t *testing.T) {
	cr := NewCellRunner()

	server, client := net.Pipe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- cr.handleConnection(server, "render")
	}()

	// Send display output.
	if err := WriteMessage(client, ProtocolMessage{
		Type: MsgDisplay,
		Mime: "text/html",
		Data: []byte("<p>chart output</p>"),
	}); err != nil {
		t.Fatal(err)
	}

	// Send another display.
	if err := WriteMessage(client, ProtocolMessage{
		Type: MsgDisplay,
		Mime: "image/svg+xml",
		Data: []byte("<svg></svg>"),
	}); err != nil {
		t.Fatal(err)
	}

	// Send result to end.
	if err := WriteMessage(client, ProtocolMessage{
		Type: MsgResult,
		Cell: "render",
		Data: []byte("render-result"),
	}); err != nil {
		t.Fatal(err)
	}

	client.Close()
	if err := <-errCh; err != nil {
		t.Fatalf("handleConnection: %v", err)
	}

	displays := cr.GetDisplays()
	if len(displays) != 2 {
		t.Fatalf("displays = %d, want 2", len(displays))
	}
	if displays[0].Mime != "text/html" {
		t.Errorf("displays[0].Mime = %q", displays[0].Mime)
	}
	if displays[0].Data != "<p>chart output</p>" {
		t.Errorf("displays[0].Data = %q", displays[0].Data)
	}
	if displays[1].Mime != "image/svg+xml" {
		t.Errorf("displays[1].Mime = %q", displays[1].Mime)
	}
}

func TestHandleConnectionCellError(t *testing.T) {
	cr := NewCellRunner()

	server, client := net.Pipe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- cr.handleConnection(server, "broken")
	}()

	// Cell reports an error.
	if err := WriteMessage(client, ProtocolMessage{
		Type: MsgError,
		Err:  "runtime panic: index out of range",
	}); err != nil {
		t.Fatal(err)
	}

	client.Close()
	err := <-errCh
	if err == nil {
		t.Fatal("expected error from handleConnection")
	}
	if got := err.Error(); got != "cell error: runtime panic: index out of range" {
		t.Errorf("error = %q", got)
	}
}

func TestHandleConnectionEOF(t *testing.T) {
	cr := NewCellRunner()

	server, client := net.Pipe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- cr.handleConnection(server, "early")
	}()

	// Close without sending anything — simulates cell crash.
	client.Close()
	if err := <-errCh; err != nil {
		t.Fatalf("handleConnection should return nil on EOF, got: %v", err)
	}
}

func TestHandleConnectionUnknownMessageType(t *testing.T) {
	cr := NewCellRunner()

	server, client := net.Pipe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- cr.handleConnection(server, "test")
	}()

	// Send an unknown message type.
	if err := WriteMessage(client, ProtocolMessage{
		Type: "unknown_type",
	}); err != nil {
		t.Fatal(err)
	}

	client.Close()
	err := <-errCh
	if err == nil {
		t.Fatal("expected error for unknown message type")
	}
}

func TestHandleConnectionMultipleUpstreams(t *testing.T) {
	cr := NewCellRunner()
	cr.SetCacheEntry("a", []byte("data-a"))
	cr.SetCacheEntry("b", []byte("data-b"))

	server, client := net.Pipe()

	errCh := make(chan error, 1)
	go func() {
		errCh <- cr.handleConnection(server, "c")
	}()

	// Fetch upstream "a".
	WriteMessage(client, ProtocolMessage{Type: MsgFetchUpstream, Cell: "a"})
	resp, _ := ReadMessage(client)
	if string(resp.Data) != "data-a" {
		t.Errorf("upstream a = %q", resp.Data)
	}

	// Fetch upstream "b".
	WriteMessage(client, ProtocolMessage{Type: MsgFetchUpstream, Cell: "b"})
	resp, _ = ReadMessage(client)
	if string(resp.Data) != "data-b" {
		t.Errorf("upstream b = %q", resp.Data)
	}

	// Send result.
	WriteMessage(client, ProtocolMessage{Type: MsgResult, Cell: "c", Data: []byte("combined")})
	client.Close()

	if err := <-errCh; err != nil {
		t.Fatalf("handleConnection: %v", err)
	}

	data, ok := cr.GetCacheEntry("c")
	if !ok || string(data) != "combined" {
		t.Errorf("cached c = %q, ok=%v", data, ok)
	}
}
