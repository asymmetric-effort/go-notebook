// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package kernel

import (
	"fmt"
	"net"
)

// SocketClient connects to the kernel's Unix socket and provides methods
// for the cell-side protocol: fetching upstream cell data, sending display
// output, and sending the final result. Generated cell main functions use
// this client when GOBOOK_SOCKET is set.
type SocketClient struct {
	conn net.Conn
}

// DialSocket connects to the kernel's Unix socket at the given path.
func DialSocket(socketPath string) (*SocketClient, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("connect to kernel socket: %w", err)
	}
	return &SocketClient{conn: conn}, nil
}

// Close closes the connection to the kernel.
func (sc *SocketClient) Close() error {
	return sc.conn.Close()
}

// FetchUpstream requests the serialized data for an upstream cell from the
// kernel. The kernel looks up the cell in its cache and returns the bytes.
func (sc *SocketClient) FetchUpstream(cellName string) ([]byte, error) {
	err := WriteMessage(sc.conn, ProtocolMessage{
		Type: MsgFetchUpstream,
		Cell: cellName,
	})
	if err != nil {
		return nil, fmt.Errorf("send fetch_upstream: %w", err)
	}

	resp, err := ReadMessage(sc.conn)
	if err != nil {
		return nil, fmt.Errorf("read upstream response: %w", err)
	}

	switch resp.Type {
	case MsgUpstreamData:
		return resp.Data, nil
	case MsgError:
		return nil, fmt.Errorf("kernel error: %s", resp.Err)
	default:
		return nil, fmt.Errorf("unexpected response type: %s", resp.Type)
	}
}

// SendDisplay sends a display output to the kernel for streaming to the
// user interface (e.g., HTML chart, image, text).
func (sc *SocketClient) SendDisplay(mime string, data []byte) error {
	return WriteMessage(sc.conn, ProtocolMessage{
		Type: MsgDisplay,
		Mime: mime,
		Data: data,
	})
}

// SendResult sends the final serialized cell result to the kernel and
// signals that the cell has completed.
func (sc *SocketClient) SendResult(cellName string, data []byte) error {
	return WriteMessage(sc.conn, ProtocolMessage{
		Type: MsgResult,
		Cell: cellName,
		Data: data,
	})
}

// SendError sends an error message to the kernel.
func (sc *SocketClient) SendError(errMsg string) error {
	return WriteMessage(sc.conn, ProtocolMessage{
		Type: MsgError,
		Err:  errMsg,
	})
}
