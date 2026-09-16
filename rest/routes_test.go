package rest

import (
	"net/http"
	"testing"

	"github.com/Khietn/gord-lib/discord"
)

func TestRoutes(t *testing.T) {
	channelID := discord.Snowflake(123456789)
	guildID := discord.Snowflake(987654321)
	appID := discord.Snowflake(1122334455)

	// CreateMessage
	rMsg := RouteCreateMessage(channelID)
	if rMsg.Method != http.MethodPost || rMsg.Path != "/channels/123456789/messages" {
		t.Errorf("unexpected RouteCreateMessage: %v", rMsg)
	}

	// GetChannel
	rCh := RouteGetChannel(channelID)
	if rCh.Method != http.MethodGet || rCh.Path != "/channels/123456789" {
		t.Errorf("unexpected RouteGetChannel: %v", rCh)
	}

	// GetGuild
	rGuild := RouteGetGuild(guildID)
	if rGuild.Method != http.MethodGet || rGuild.Path != "/guilds/987654321" {
		t.Errorf("unexpected RouteGetGuild: %v", rGuild)
	}

	// CreateInteractionResponse
	rInteraction := RouteCreateInteractionResponse(discord.Snowflake(111), "token123")
	if rInteraction.Method != http.MethodPost || rInteraction.Path != "/interactions/111/token123/callback" {
		t.Errorf("unexpected RouteCreateInteractionResponse: %v", rInteraction)
	}

	// CreateGlobalCommand
	rGlobal := RouteCreateGlobalCommand(appID)
	if rGlobal.Method != http.MethodPost || rGlobal.Path != "/applications/1122334455/commands" {
		t.Errorf("unexpected RouteCreateGlobalCommand: %v", rGlobal)
	}

	// CreateGuildCommand
	rGuildCmd := RouteCreateGuildCommand(appID, guildID)
	if rGuildCmd.Method != http.MethodPost || rGuildCmd.Path != "/applications/1122334455/guilds/987654321/commands" {
		t.Errorf("unexpected RouteCreateGuildCommand: %v", rGuildCmd)
	}

	// GetGatewayBot
	rGateway := RouteGetGatewayBot()
	if rGateway.Method != http.MethodGet || rGateway.Path != "/gateway/bot" {
		t.Errorf("unexpected RouteGetGatewayBot: %v", rGateway)
	}
}
