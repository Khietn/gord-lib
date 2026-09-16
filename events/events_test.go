package events

import (
	"testing"
)

func TestEventNames(t *testing.T) {
	tests := []struct {
		event    Event
		expected string
	}{
		{Ready{}, "READY"},
		{MessageCreate{}, "MESSAGE_CREATE"},
		{MessageUpdate{}, "MESSAGE_UPDATE"},
		{MessageDelete{}, "MESSAGE_DELETE"},
		{InteractionCreate{}, "INTERACTION_CREATE"},
		{GuildCreate{}, "GUILD_CREATE"},
		{GuildDelete{}, "GUILD_DELETE"},
	}

	for _, tt := range tests {
		if tt.event.EventName() != tt.expected {
			t.Errorf("expected event name %q, got %q", tt.expected, tt.event.EventName())
		}
	}
}
