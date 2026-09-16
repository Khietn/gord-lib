package gateway

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Khietn/gord-lib/discord"
	"github.com/Khietn/gord-lib/events"
)

type mockConn struct {
	readCh  chan []byte
	written [][]byte
	closed  bool
}

func newMockConn() *mockConn {
	return &mockConn{
		readCh: make(chan []byte, 20),
	}
}

func (m *mockConn) Read(ctx context.Context) (MessageType, []byte, error) {
	select {
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	case data, ok := <-m.readCh:
		if !ok {
			return 0, nil, context.Canceled
		}
		return TextMessage, data, nil
	}
}

func (m *mockConn) Write(ctx context.Context, messageType MessageType, data []byte) error {
	m.written = append(m.written, data)
	return nil
}

func (m *mockConn) Close(code int, reason string) error {
	m.closed = true
	return nil
}

type mockDialer struct {
	conn *mockConn
}

func (d *mockDialer) Dial(ctx context.Context, url string) (Conn, error) {
	return d.conn, nil
}

func TestGateway_LifecycleAndPayloads(t *testing.T) {
	dispatcher := events.NewDispatcher(events.WithWorkerCount(4))
	defer dispatcher.Close()

	var msgReceived string
	var readyReceived bool
	var wg sync.WaitGroup
	wg.Add(2)

	events.Register(dispatcher, func(ctx context.Context, e events.Ready) {
		readyReceived = true
		wg.Done()
	})

	events.Register(dispatcher, func(ctx context.Context, e events.MessageCreate) {
		msgReceived = e.Content
		wg.Done()
	})

	conn := newMockConn()
	gw := New(Config{
		Token:      "mock-token",
		Intents:    discord.IntentsDefault,
		Dialer:     &mockDialer{conn: conn},
		Dispatcher: dispatcher,
	})

	if gw.State() != StateDisconnected {
		t.Fatalf("expected state Disconnected, got %s", gw.State())
	}

	ctx := context.Background()
	if err := gw.Open(ctx, "wss://gateway.discord.gg"); err != nil {
		t.Fatalf("failed to open gateway: %v", err)
	}

	if gw.State() != StateHandshaking {
		t.Fatalf("expected state Handshaking, got %s", gw.State())
	}

	// 1. Send OpcodeHello frame
	conn.readCh <- []byte(`{"op":10,"d":{"heartbeat_interval":40000}}`)

	// 2. Send READY event frame
	conn.readCh <- []byte(`{"op":0,"t":"READY","s":1,"d":{"v":10,"session_id":"sess123","user":{"id":"123","username":"TestBot"},"resume_gateway_url":"wss://resume.discord.gg"}}`)

	// 3. Send MESSAGE_CREATE event frame
	conn.readCh <- []byte(`{"op":0,"t":"MESSAGE_CREATE","s":2,"d":{"id":"999","channel_id":"888","content":"Hello Gateway!"}}`)

	// 4. Send OpcodeHeartbeat (requesting immediate heartbeat)
	conn.readCh <- []byte(`{"op":1}`)

	// 5. Send OpcodeHeartbeatACK
	conn.readCh <- []byte(`{"op":11}`)

	// Wait for event handlers
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for events")
	}

	if !readyReceived {
		t.Errorf("expected READY event to be dispatched")
	}
	if msgReceived != "Hello Gateway!" {
		t.Errorf("expected message 'Hello Gateway!', got %s", msgReceived)
	}
	if gw.State() != StateReady {
		t.Errorf("expected state Ready, got %s", gw.State())
	}

	// Test Nil connection error
	nilGw := New(Config{Token: "test"})
	if err := nilGw.WriteJSON(ctx, map[string]string{}); err == nil {
		t.Errorf("expected error writing to nil conn")
	}

	// Close Gateway
	_ = gw.Close()
	if gw.State() != StateClosed {
		t.Fatalf("expected state Closed, got %s", gw.State())
	}
}

func TestNewDefaultDialer(t *testing.T) {
	d := NewDefaultDialer()
	if d == nil || d.dialer == nil {
		t.Fatalf("expected initialized DefaultDialer")
	}
}
