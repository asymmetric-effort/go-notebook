// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package kernel

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

// Message types for the kernel-cell protocol over Unix sockets.
const (
	MsgFetchUpstream = "fetch_upstream" // Cell → Kernel: request upstream cell data.
	MsgUpstreamData  = "upstream_data"  // Kernel → Cell: serialized upstream cell.
	MsgDisplay       = "display"        // Cell → Kernel: display output (streaming).
	MsgResult        = "result"         // Cell → Kernel: final cell result.
	MsgError         = "error"          // Either direction: error message.
)

// ProtocolMessage is a framed message in the kernel-cell protocol.
type ProtocolMessage struct {
	Type string `json:"type"`
	Cell string `json:"cell,omitempty"` // Cell name for fetch/result.
	Mime string `json:"mime,omitempty"` // MIME type for display output.
	Data []byte `json:"data,omitempty"` // Payload (serialized cell or display content).
	Err  string `json:"error,omitempty"`
}

// WriteMessage sends a framed message over the connection.
// Frame format: 4-byte big-endian length + JSON payload.
func WriteMessage(conn net.Conn, msg ProtocolMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	length := uint32(len(payload))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return fmt.Errorf("write length: %w", err)
	}
	if _, err := conn.Write(payload); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}
	return nil
}

// ReadMessage receives a framed message from the connection.
func ReadMessage(conn net.Conn) (ProtocolMessage, error) {
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		if err == io.EOF {
			return ProtocolMessage{}, io.EOF
		}
		return ProtocolMessage{}, fmt.Errorf("read length: %w", err)
	}
	if length > 100*1024*1024 { // 100 MB max message size.
		return ProtocolMessage{}, fmt.Errorf("message too large: %d bytes", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return ProtocolMessage{}, fmt.Errorf("read payload: %w", err)
	}
	var msg ProtocolMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return ProtocolMessage{}, fmt.Errorf("unmarshal message: %w", err)
	}
	return msg, nil
}
