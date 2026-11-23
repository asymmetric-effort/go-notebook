// Copyright (c) 2026 Asymmetric Effort, LLC. MIT License.

package kernel

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"testing"
)

func TestWriteReadMessageRoundTrip(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	msg := ProtocolMessage{
		Type: MsgResult,
		Cell: "load",
		Data: []byte("hello world"),
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- WriteMessage(server, msg)
	}()

	got, err := ReadMessage(client)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	if got.Type != msg.Type {
		t.Errorf("Type = %q, want %q", got.Type, msg.Type)
	}
	if got.Cell != msg.Cell {
		t.Errorf("Cell = %q, want %q", got.Cell, msg.Cell)
	}
	if !bytes.Equal(got.Data, msg.Data) {
		t.Errorf("Data = %q, want %q", got.Data, msg.Data)
	}
}

func TestWriteReadMessageAllTypes(t *testing.T) {
	tests := []struct {
		name string
		msg  ProtocolMessage
	}{
		{
			name: "fetch_upstream",
			msg:  ProtocolMessage{Type: MsgFetchUpstream, Cell: "load"},
		},
		{
			name: "upstream_data",
			msg:  ProtocolMessage{Type: MsgUpstreamData, Cell: "load", Data: []byte(`{"val":42}`)},
		},
		{
			name: "display",
			msg:  ProtocolMessage{Type: MsgDisplay, Mime: "text/html", Data: []byte("<p>hello</p>")},
		},
		{
			name: "result",
			msg:  ProtocolMessage{Type: MsgResult, Cell: "transform", Data: []byte("serialized")},
		},
		{
			name: "error",
			msg:  ProtocolMessage{Type: MsgError, Err: "something went wrong"},
		},
		{
			name: "empty_data",
			msg:  ProtocolMessage{Type: MsgResult, Cell: "empty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, client := net.Pipe()
			defer server.Close()
			defer client.Close()

			errCh := make(chan error, 1)
			go func() {
				errCh <- WriteMessage(server, tt.msg)
			}()

			got, err := ReadMessage(client)
			if err != nil {
				t.Fatalf("ReadMessage: %v", err)
			}
			if err := <-errCh; err != nil {
				t.Fatalf("WriteMessage: %v", err)
			}

			if got.Type != tt.msg.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.msg.Type)
			}
			if got.Cell != tt.msg.Cell {
				t.Errorf("Cell = %q, want %q", got.Cell, tt.msg.Cell)
			}
			if got.Mime != tt.msg.Mime {
				t.Errorf("Mime = %q, want %q", got.Mime, tt.msg.Mime)
			}
			if !bytes.Equal(got.Data, tt.msg.Data) {
				t.Errorf("Data = %q, want %q", got.Data, tt.msg.Data)
			}
			if got.Err != tt.msg.Err {
				t.Errorf("Err = %q, want %q", got.Err, tt.msg.Err)
			}
		})
	}
}

func TestMultipleMessages(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	msgs := []ProtocolMessage{
		{Type: MsgFetchUpstream, Cell: "a"},
		{Type: MsgUpstreamData, Cell: "a", Data: []byte("data-a")},
		{Type: MsgDisplay, Mime: "text/plain", Data: []byte("output")},
		{Type: MsgResult, Cell: "b", Data: []byte("result-b")},
	}

	errCh := make(chan error, 1)
	go func() {
		for _, m := range msgs {
			if err := WriteMessage(server, m); err != nil {
				errCh <- err
				return
			}
		}
		errCh <- nil
	}()

	for i, want := range msgs {
		got, err := ReadMessage(client)
		if err != nil {
			t.Fatalf("ReadMessage[%d]: %v", i, err)
		}
		if got.Type != want.Type {
			t.Errorf("[%d] Type = %q, want %q", i, got.Type, want.Type)
		}
		if got.Cell != want.Cell {
			t.Errorf("[%d] Cell = %q, want %q", i, got.Cell, want.Cell)
		}
		if !bytes.Equal(got.Data, want.Data) {
			t.Errorf("[%d] Data mismatch", i)
		}
	}

	if err := <-errCh; err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
}

func TestReadMessageEOF(t *testing.T) {
	server, client := net.Pipe()
	server.Close() // Close immediately so reader gets EOF.

	_, err := ReadMessage(client)
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
	client.Close()
}

func TestReadMessageTooLarge(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Write a length header that exceeds the 100MB limit.
	go func() {
		binary.Write(server, binary.BigEndian, uint32(101*1024*1024))
	}()

	_, err := ReadMessage(client)
	if err == nil {
		t.Fatal("expected error for oversized message")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("too large")) {
		t.Errorf("error = %q, want 'too large'", err)
	}
}

func TestReadMessageTruncatedPayload(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	// Write a valid length header but then close before sending full payload.
	go func() {
		binary.Write(server, binary.BigEndian, uint32(100))
		server.Write([]byte("short")) // Only 5 bytes instead of 100.
		server.Close()
	}()

	_, err := ReadMessage(client)
	if err == nil {
		t.Fatal("expected error for truncated payload")
	}
}

func TestReadMessageInvalidJSON(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Write valid framing but invalid JSON payload.
	go func() {
		payload := []byte("not valid json {{{")
		binary.Write(server, binary.BigEndian, uint32(len(payload)))
		server.Write(payload)
	}()

	_, err := ReadMessage(client)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("unmarshal")) {
		t.Errorf("error = %q, want unmarshal error", err)
	}
}

func TestWriteMessageLargePayload(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// 1MB of data to test larger payloads.
	bigData := make([]byte, 1024*1024)
	for i := range bigData {
		bigData[i] = byte(i % 256)
	}

	msg := ProtocolMessage{
		Type: MsgResult,
		Cell: "big",
		Data: bigData,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- WriteMessage(server, msg)
	}()

	got, err := ReadMessage(client)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	if !bytes.Equal(got.Data, bigData) {
		t.Errorf("large payload mismatch: got %d bytes, want %d", len(got.Data), len(bigData))
	}
}

func TestProtocolMessageJSON(t *testing.T) {
	msg := ProtocolMessage{
		Type: MsgDisplay,
		Mime: "image/png",
		Data: []byte{0x89, 0x50, 0x4e, 0x47},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	var got ProtocolMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if got.Type != msg.Type {
		t.Errorf("Type = %q, want %q", got.Type, msg.Type)
	}
	if !bytes.Equal(got.Data, msg.Data) {
		t.Errorf("Data mismatch after JSON round-trip")
	}
}

func TestMessageConstants(t *testing.T) {
	// Verify the message type constants are distinct.
	types := map[string]bool{
		MsgFetchUpstream: true,
		MsgUpstreamData:  true,
		MsgDisplay:       true,
		MsgResult:        true,
		MsgError:         true,
	}
	if len(types) != 5 {
		t.Error("message type constants are not all distinct")
	}
}
