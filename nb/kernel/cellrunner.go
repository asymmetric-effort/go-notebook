// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package kernel

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// CellRunner manages per-cell process execution over Unix sockets.
// Each cell runs in its own process and communicates with the kernel
// through the socket protocol.
type CellRunner struct {
	cache    map[string][]byte // Cached serialized cell results by name.
	displays []DisplayOutput   // Display output received from cells.
	mu       sync.Mutex
}

// DisplayOutput holds a display entry received from a cell.
type DisplayOutput struct {
	Mime string
	Data string
}

// NewCellRunner creates a cell runner with an empty cache.
func NewCellRunner() *CellRunner {
	return &CellRunner{
		cache: make(map[string][]byte),
	}
}

// SetCacheEntry stores a serialized cell result in the cache.
func (cr *CellRunner) SetCacheEntry(name string, data []byte) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.cache[name] = data
}

// GetCacheEntry retrieves a cached cell result.
func (cr *CellRunner) GetCacheEntry(name string) ([]byte, bool) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	data, ok := cr.cache[name]
	return data, ok
}

// GetDisplays returns all display output received from cells.
func (cr *CellRunner) GetDisplays() []DisplayOutput {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	result := make([]DisplayOutput, len(cr.displays))
	copy(result, cr.displays)
	return result
}

// ClearDisplays resets the display output buffer.
func (cr *CellRunner) ClearDisplays() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.displays = nil
}

// RunCellOverSocket starts a Unix socket listener, launches the cell
// process, handles the protocol exchange, and returns the cell result.
// The cell process connects to the socket, fetches upstream results,
// streams display output, and sends its final result.
func (cr *CellRunner) RunCellOverSocket(ctx context.Context, cellName string, pkgDir string) (CellResult, error) {
	// Create a temporary Unix socket.
	socketDir, err := os.MkdirTemp("", "gobook-sock-*")
	if err != nil {
		return CellResult{}, fmt.Errorf("create socket dir: %w", err)
	}
	defer os.RemoveAll(socketDir)

	socketPath := filepath.Join(socketDir, "cell.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return CellResult{}, fmt.Errorf("listen on socket: %w", err)
	}
	defer listener.Close()

	// Accept the connection in a goroutine. The cell process will connect
	// after it starts up.
	type connResult struct {
		err error
	}
	connDone := make(chan connResult, 1)

	go func() {
		// Use a deadline tied to the context so we don't block forever
		// if the cell process never connects.
		conn, err := listener.Accept()
		if err != nil {
			// If context was cancelled, the listener was closed.
			connDone <- connResult{err: fmt.Errorf("accept: %w", err)}
			return
		}
		connDone <- connResult{err: cr.handleConnection(conn, cellName)}
	}()

	// Launch the cell process with the socket path as an env var.
	result, runErr := GoRunWithEnv(ctx, pkgDir, []string{
		"GOBOOK_SOCKET=" + socketPath,
	})

	// Close the listener to unblock Accept if the process exited without connecting.
	listener.Close()

	// Wait for the connection handler to finish.
	cr2 := <-connDone

	// Populate display output from the socket protocol.
	displays := cr.GetDisplays()
	for _, d := range displays {
		result.Displays = append(result.Displays, DisplayEntry{
			Mime: d.Mime,
			Data: d.Data,
		})
	}

	// Populate serialized data from cache if the cell sent a result.
	if data, ok := cr.GetCacheEntry(cellName); ok {
		result.Data = data
	}

	// Return the process error if the cell didn't connect or had a protocol error.
	if runErr != nil {
		return result, runErr
	}
	if cr2.err != nil {
		return result, fmt.Errorf("socket protocol: %w", cr2.err)
	}

	return result, nil
}

// handleConnection processes messages from a connected cell process.
func (cr *CellRunner) handleConnection(conn net.Conn, cellName string) error {
	defer conn.Close()

	for {
		msg, err := ReadMessage(conn)
		if err != nil {
			if err == io.EOF {
				return nil // Cell process exited.
			}
			return fmt.Errorf("read message: %w", err)
		}

		switch msg.Type {
		case MsgFetchUpstream:
			data, ok := cr.GetCacheEntry(msg.Cell)
			if !ok {
				WriteMessage(conn, ProtocolMessage{
					Type: MsgError,
					Err:  fmt.Sprintf("upstream cell %q not in cache", msg.Cell),
				})
				continue
			}
			WriteMessage(conn, ProtocolMessage{
				Type: MsgUpstreamData,
				Cell: msg.Cell,
				Data: data,
			})

		case MsgDisplay:
			cr.mu.Lock()
			cr.displays = append(cr.displays, DisplayOutput{
				Mime: msg.Mime,
				Data: string(msg.Data),
			})
			cr.mu.Unlock()

		case MsgResult:
			cr.SetCacheEntry(cellName, msg.Data)
			return nil // Cell completed.

		case MsgError:
			return fmt.Errorf("cell error: %s", msg.Err)

		default:
			return fmt.Errorf("unknown message type: %s", msg.Type)
		}
	}
}
