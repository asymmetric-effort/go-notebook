// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package kernel

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// shortSockDir creates a short temp directory suitable for Unix socket paths.
// macOS has a ~104-byte limit on socket paths; t.TempDir() can exceed this.
func shortSockDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sock-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestDialSocketAndClose(t *testing.T) {
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// Accept in background.
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, _ := listener.Accept()
		accepted <- conn
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	conn := <-accepted
	defer conn.Close()
}

func TestDialSocketNotExist(t *testing.T) {
	_, err := DialSocket("/nonexistent/path/test.sock")
	if err == nil {
		t.Fatal("expected error for nonexistent socket")
	}
}

func TestSocketClientFetchUpstream(t *testing.T) {
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, _ := listener.Accept()
		accepted <- conn
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	serverConn := <-accepted
	defer serverConn.Close()

	// Server responds to fetch_upstream in background.
	go func() {
		msg, _ := ReadMessage(serverConn)
		if msg.Type == MsgFetchUpstream && msg.Cell == "load" {
			WriteMessage(serverConn, ProtocolMessage{
				Type: MsgUpstreamData,
				Cell: "load",
				Data: []byte("load-data"),
			})
		}
	}()

	data, err := client.FetchUpstream("load")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "load-data" {
		t.Errorf("data = %q, want %q", data, "load-data")
	}
}

func TestSocketClientFetchUpstreamError(t *testing.T) {
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, _ := listener.Accept()
		accepted <- conn
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	serverConn := <-accepted
	defer serverConn.Close()

	// Server responds with error.
	go func() {
		ReadMessage(serverConn)
		WriteMessage(serverConn, ProtocolMessage{
			Type: MsgError,
			Err:  "cell not found",
		})
	}()

	_, err = client.FetchUpstream("missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSocketClientSendDisplay(t *testing.T) {
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, _ := listener.Accept()
		accepted <- conn
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	serverConn := <-accepted
	defer serverConn.Close()

	if err := client.SendDisplay("text/html", []byte("<p>hello</p>")); err != nil {
		t.Fatal(err)
	}

	msg, err := ReadMessage(serverConn)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != MsgDisplay {
		t.Errorf("Type = %q, want %q", msg.Type, MsgDisplay)
	}
	if msg.Mime != "text/html" {
		t.Errorf("Mime = %q", msg.Mime)
	}
	if string(msg.Data) != "<p>hello</p>" {
		t.Errorf("Data = %q", msg.Data)
	}
}

func TestSocketClientSendResult(t *testing.T) {
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, _ := listener.Accept()
		accepted <- conn
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	serverConn := <-accepted
	defer serverConn.Close()

	if err := client.SendResult("transform", []byte("result-data")); err != nil {
		t.Fatal(err)
	}

	msg, err := ReadMessage(serverConn)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != MsgResult {
		t.Errorf("Type = %q, want %q", msg.Type, MsgResult)
	}
	if msg.Cell != "transform" {
		t.Errorf("Cell = %q", msg.Cell)
	}
	if string(msg.Data) != "result-data" {
		t.Errorf("Data = %q", msg.Data)
	}
}

func TestSocketClientSendError(t *testing.T) {
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, _ := listener.Accept()
		accepted <- conn
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	serverConn := <-accepted
	defer serverConn.Close()

	if err := client.SendError("something failed"); err != nil {
		t.Fatal(err)
	}

	msg, err := ReadMessage(serverConn)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != MsgError {
		t.Errorf("Type = %q, want %q", msg.Type, MsgError)
	}
	if msg.Err != "something failed" {
		t.Errorf("Err = %q", msg.Err)
	}
}

func TestSocketClientFullProtocol(t *testing.T) {
	// End-to-end: client fetches upstream, sends display, sends result.
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	cr := NewCellRunner()
	cr.SetCacheEntry("load", []byte("load-serialized"))

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, _ := listener.Accept()
		accepted <- conn
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}

	serverConn := <-accepted

	// Run handleConnection on the server side.
	handleDone := make(chan error, 1)
	go func() {
		handleDone <- cr.handleConnection(serverConn, "transform")
	}()

	// Client: fetch upstream "load".
	data, err := client.FetchUpstream("load")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "load-serialized" {
		t.Errorf("upstream data = %q", data)
	}

	// Client: send display.
	if err := client.SendDisplay("text/plain", []byte("step 1 done")); err != nil {
		t.Fatal(err)
	}

	// Client: send result.
	if err := client.SendResult("transform", []byte("transform-result")); err != nil {
		t.Fatal(err)
	}
	client.Close()

	// Wait for handleConnection.
	if err := <-handleDone; err != nil {
		t.Fatalf("handleConnection: %v", err)
	}

	// Verify result cached.
	cached, ok := cr.GetCacheEntry("transform")
	if !ok {
		t.Fatal("transform not cached")
	}
	if string(cached) != "transform-result" {
		t.Errorf("cached = %q", cached)
	}

	// Verify display captured.
	displays := cr.GetDisplays()
	if len(displays) != 1 {
		t.Fatalf("displays = %d, want 1", len(displays))
	}
	if displays[0].Mime != "text/plain" || displays[0].Data != "step 1 done" {
		t.Errorf("display = %+v", displays[0])
	}
}

func TestDialSocketOverRealUnixSocket(t *testing.T) {
	// Test the full Dial path with a real temporary socket file.
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "real.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// Verify socket file exists.
	if _, err := os.Stat(sockPath); err != nil {
		t.Fatalf("socket file not found: %v", err)
	}

	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			conn.Close()
		}
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	client.Close()
}

func TestSocketClientFetchUpstreamWriteError(t *testing.T) {
	// Cover FetchUpstream "send fetch_upstream" error path.
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			conn.Close() // Close immediately so client write fails.
		}
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	// Wait for server to close the connection.
	ReadMessage(client.conn) // Will fail, but that's ok — just ensures server closed.

	_, err = client.FetchUpstream("test")
	client.Close()
	if err == nil {
		t.Fatal("expected error from FetchUpstream on closed conn")
	}
}

func TestSocketClientFetchUpstreamReadError(t *testing.T) {
	// Cover FetchUpstream "read upstream response" error path.
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			// Read the fetch_upstream request but close without responding.
			ReadMessage(conn)
			conn.Close()
		}
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	_, err = client.FetchUpstream("test")
	if err == nil {
		t.Fatal("expected error from FetchUpstream when server closes")
	}
}

func TestSocketClientFetchUpstreamUnexpectedType(t *testing.T) {
	// Cover FetchUpstream "unexpected response type" error path.
	dir := shortSockDir(t)
	sockPath := filepath.Join(dir, "test.sock")

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			defer conn.Close()
			ReadMessage(conn) // Read the request.
			// Reply with an unexpected type.
			WriteMessage(conn, ProtocolMessage{
				Type: MsgDisplay,
				Mime: "text/plain",
				Data: []byte("wrong type"),
			})
		}
	}()

	client, err := DialSocket(sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	_, err = client.FetchUpstream("test")
	if err == nil {
		t.Fatal("expected error for unexpected response type")
	}
	if !strings.Contains(err.Error(), "unexpected response type") {
		t.Errorf("error = %q, want 'unexpected response type'", err)
	}
}
