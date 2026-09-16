package gateway

import (
	"context"
)

// MessageType identifies the type of a WebSocket frame.
type MessageType int

const (
	TextMessage   MessageType = 1
	BinaryMessage MessageType = 2
)

// Conn represents an abstract WebSocket connection interface.
// This allows gord-lib to be completely decoupled from specific WebSocket implementations
// and enables 100% mockable, zero-allocation testing.
type Conn interface {
	// Read reads the next incoming message frame.
	Read(ctx context.Context) (MessageType, []byte, error)
	// Write sends a message frame to the remote endpoint.
	Write(ctx context.Context, messageType MessageType, data []byte) error
	// Close gracefully terminates the connection.
	Close(code int, reason string) error
}

// Dialer abstracts opening a WebSocket connection to the Discord Gateway.
type Dialer interface {
	Dial(ctx context.Context, endpointURL string) (Conn, error)
}
