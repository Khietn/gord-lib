package gateway

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// GorillaConn wraps a *websocket.Conn to conform to our Conn interface.
type GorillaConn struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
}

// Read reads the next frame from the Gorilla WebSocket.
func (c *GorillaConn) Read(ctx context.Context) (MessageType, []byte, error) {
	msgType, data, err := c.conn.ReadMessage()
	if err != nil {
		return 0, nil, err
	}
	return MessageType(msgType), data, nil
}

// Write sends a message frame to the Gorilla WebSocket thread-safely.
func (c *GorillaConn) Write(ctx context.Context, messageType MessageType, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.conn.WriteMessage(int(messageType), data)
}

// Close gracefully sends a close message and closes the underlying network connection.
func (c *GorillaConn) Close(code int, reason string) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	_ = c.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, reason),
		time.Now().Add(time.Second),
	)
	return c.conn.Close()
}

// DefaultDialer connects to Discord WebSocket gateway using Gorilla WebSocket.
type DefaultDialer struct {
	dialer *websocket.Dialer
}

// NewDefaultDialer creates a new default WebSocket dialer.
func NewDefaultDialer() *DefaultDialer {
	return &DefaultDialer{
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// Dial connects to the gateway WebSocket URL.
func (d *DefaultDialer) Dial(ctx context.Context, endpointURL string) (Conn, error) {
	conn, _, err := d.dialer.DialContext(ctx, endpointURL, nil)
	if err != nil {
		return nil, err
	}
	return &GorillaConn{conn: conn}, nil
}
