package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Khietn/gord-lib/discord"
)

func TestClient_OptionsAndEndpoints(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check standard headers
		if r.Header.Get("Authorization") != "Bot test-token" {
			http.Error(w, `{"code": 40001, "message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if r.Header.Get("User-Agent") != "CustomUA" {
			http.Error(w, `{"code": 40000, "message": "Bad User Agent"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/gateway/bot":
			_ = json.NewEncoder(w).Encode(GatewayBotResponse{
				URL:    "wss://gateway.discord.gg",
				Shards: 1,
			})

		case "/channels/12345":
			_ = json.NewEncoder(w).Encode(discord.Channel{
				ID:   discord.Snowflake(12345),
				Name: "general",
			})

		case "/channels/12345/messages":
			_ = json.NewEncoder(w).Encode(discord.Message{
				ID:        discord.Snowflake(999),
				ChannelID: discord.Snowflake(12345),
				Content:   "Mock message reply",
			})

		case "/guilds/67890":
			_ = json.NewEncoder(w).Encode(discord.Guild{
				ID:   discord.Snowflake(67890),
				Name: "Test Guild",
			})

		case "/interactions/111/tok/callback":
			w.WriteHeader(http.StatusNoContent)

		case "/error":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(DiscordError{
				Code:    10003,
				Message: "Unknown Channel",
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	customHTTP := &http.Client{Timeout: 5 * time.Second}
	customRL := NewRateLimiter()

	client := NewClient("test-token",
		WithBaseURL(server.URL),
		WithUserAgent("CustomUA"),
		WithHTTPClient(customHTTP),
		WithRateLimiter(customRL),
	)

	ctx := context.Background()

	// 1. GetGatewayBot
	gw, err := client.GetGatewayBot(ctx)
	if err != nil {
		t.Fatalf("GetGatewayBot failed: %v", err)
	}
	if gw.URL != "wss://gateway.discord.gg" {
		t.Errorf("expected URL wss://gateway.discord.gg, got %s", gw.URL)
	}

	// 2. GetChannel
	ch, err := client.GetChannel(ctx, discord.Snowflake(12345))
	if err != nil {
		t.Fatalf("GetChannel failed: %v", err)
	}
	if ch.Name != "general" {
		t.Errorf("expected channel name 'general', got %s", ch.Name)
	}

	// 3. CreateMessage
	msg, err := client.CreateMessage(ctx, discord.Snowflake(12345), discord.MessageCreatePayload{
		Content: "Hello",
	})
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}
	if msg.Content != "Mock message reply" {
		t.Errorf("expected reply content, got %s", msg.Content)
	}

	// 4. GetGuild
	guild, err := client.GetGuild(ctx, discord.Snowflake(67890))
	if err != nil {
		t.Fatalf("GetGuild failed: %v", err)
	}
	if guild.Name != "Test Guild" {
		t.Errorf("expected guild name 'Test Guild', got %s", guild.Name)
	}

	// 5. CreateInteractionResponse
	err = client.CreateInteractionResponse(ctx, discord.Snowflake(111), "tok", discord.InteractionResponse{
		Type: discord.InteractionCallbackTypePong,
	})
	if err != nil {
		t.Fatalf("CreateInteractionResponse failed: %v", err)
	}

	// 6. DiscordError test
	errRoute := &Route{
		Method:     http.MethodGet,
		Path:       "/error",
		BucketKey:  "/error",
		MajorParam: "global",
	}
	err = client.Do(ctx, errRoute, nil, nil)
	if err == nil {
		t.Fatalf("expected error from /error route, got nil")
	}
	discErr, ok := err.(*DiscordError)
	if !ok {
		t.Fatalf("expected *DiscordError, got %T: %v", err, err)
	}
	if discErr.Code != 10003 {
		t.Errorf("expected code 10003, got %d", discErr.Code)
	}
	if discErr.Error() == "" {
		t.Errorf("expected non-empty Error() string")
	}
}
