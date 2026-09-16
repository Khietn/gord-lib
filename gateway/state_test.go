package gateway

import (
	"testing"
)

func TestState_String(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateDisconnected, "Disconnected"},
		{StateConnecting, "Connecting"},
		{StateHandshaking, "Handshaking"},
		{StateIdentifying, "Identifying"},
		{StateResuming, "Resuming"},
		{StateReady, "Ready"},
		{StateReconnecting, "Reconnecting"},
		{StateClosed, "Closed"},
		{State(999), "Unknown"},
	}

	for _, tt := range tests {
		if tt.state.String() != tt.expected {
			t.Errorf("expected state %d to be %q, got %q", tt.state, tt.expected, tt.state.String())
		}
	}
}
