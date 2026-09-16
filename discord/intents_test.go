package discord

import (
	"testing"
)

func TestGatewayIntents(t *testing.T) {
	intents := IntentGuilds | IntentGuildMessages

	if !intents.Has(IntentGuilds) {
		t.Errorf("expected intents to have IntentGuilds")
	}
	if !intents.Has(IntentGuildMessages) {
		t.Errorf("expected intents to have IntentGuildMessages")
	}
	if intents.Has(IntentMessageContent) {
		t.Errorf("expected intents to NOT have IntentMessageContent")
	}

	// Test Add
	updated := intents.Add(IntentMessageContent, IntentGuildMembers)
	if !updated.Has(IntentMessageContent) {
		t.Errorf("expected updated intents to have IntentMessageContent")
	}
	if !updated.Has(IntentGuildMembers) {
		t.Errorf("expected updated intents to have IntentGuildMembers")
	}

	// Test Remove
	removed := updated.Remove(IntentGuildMessages)
	if removed.Has(IntentGuildMessages) {
		t.Errorf("expected removed intents to NOT have IntentGuildMessages")
	}
	if !removed.Has(IntentGuilds) {
		t.Errorf("expected removed intents to still have IntentGuilds")
	}

	// Test Predefined groups
	if !IntentsAll.Has(IntentsPrivileged) {
		t.Errorf("IntentsAll must contain IntentsPrivileged")
	}
	if !IntentsAll.Has(IntentsDefault) {
		t.Errorf("IntentsAll must contain IntentsDefault")
	}
	if IntentsDefault.Has(IntentMessageContent) {
		t.Errorf("IntentsDefault must not contain privileged IntentMessageContent")
	}
}
