package discord

import (
	"testing"
)

func TestMessageBuilder(t *testing.T) {
	embed, err := NewEmbedBuilder().
		SetTitle("Message Embed").
		Build()
	if err != nil {
		t.Fatalf("failed to build embed: %v", err)
	}

	msgID := Snowflake(111222333)
	channelID := Snowflake(444555666)

	payload := NewMessageBuilder().
		SetContent("Test Message Content").
		SetTTS(true).
		AddEmbed(embed).
		SetFlags(MessageFlagSuppressNotifications).
		ReplyTo(msgID, channelID).
		Build()

	if payload.Content != "Test Message Content" {
		t.Errorf("expected content 'Test Message Content', got %s", payload.Content)
	}
	if !payload.TTS {
		t.Errorf("expected TTS to be true")
	}
	if len(payload.Embeds) != 1 || payload.Embeds[0].Title != "Message Embed" {
		t.Errorf("expected 1 embed with title 'Message Embed'")
	}
	if payload.Flags != MessageFlagSuppressNotifications {
		t.Errorf("expected flag MessageFlagSuppressNotifications, got %d", payload.Flags)
	}
	if payload.MessageReference == nil || payload.MessageReference.MessageID != msgID || payload.MessageReference.ChannelID != channelID {
		t.Errorf("expected MessageReference to match reply target")
	}
}

func TestSnowflake_EdgeCases(t *testing.T) {
	var zero Snowflake
	if zero.IsValid() {
		t.Errorf("zero snowflake should not be valid")
	}

	sf := Snowflake(999999)
	if !sf.IsValid() {
		t.Errorf("non-zero snowflake should be valid")
	}

	// Null unmarshal
	var nullSf Snowflake
	if err := nullSf.UnmarshalJSON([]byte("null")); err != nil {
		t.Errorf("null unmarshal should succeed: %v", err)
	}
	if nullSf != 0 {
		t.Errorf("expected 0, got %v", nullSf)
	}

	// Invalid string
	var badSf Snowflake
	if err := badSf.UnmarshalJSON([]byte(`"not-a-number"`)); err == nil {
		t.Errorf("expected error unmarshaling non-numeric string into snowflake")
	}
}
