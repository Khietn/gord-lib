package rest

import (
	"fmt"
	"net/http"

	"github.com/Khietn/gord-lib/discord"
)

// Route represents an endpoint route definition with its HTTP method,
// formatted path, and major parameter for rate-limiting bucket isolation.
type Route struct {
	Method     string
	Path       string
	BucketKey  string
	MajorParam string
}

// RouteCreateMessage creates a message in a channel.
func RouteCreateMessage(channelID discord.Snowflake) *Route {
	return &Route{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("/channels/%s/messages", channelID),
		BucketKey:  "/channels/:channel_id/messages",
		MajorParam: channelID.String(),
	}
}

// RouteGetChannel retrieves a channel by ID.
func RouteGetChannel(channelID discord.Snowflake) *Route {
	return &Route{
		Method:     http.MethodGet,
		Path:       fmt.Sprintf("/channels/%s", channelID),
		BucketKey:  "/channels/:channel_id",
		MajorParam: channelID.String(),
	}
}

// RouteGetGuild retrieves a guild by ID.
func RouteGetGuild(guildID discord.Snowflake) *Route {
	return &Route{
		Method:     http.MethodGet,
		Path:       fmt.Sprintf("/guilds/%s", guildID),
		BucketKey:  "/guilds/:guild_id",
		MajorParam: guildID.String(),
	}
}

// RouteCreateInteractionResponse sends a response to an interaction callback.
func RouteCreateInteractionResponse(interactionID discord.Snowflake, token string) *Route {
	return &Route{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("/interactions/%s/%s/callback", interactionID, token),
		BucketKey:  "/interactions/:id/:token/callback",
		MajorParam: interactionID.String(),
	}
}

// RouteCreateGlobalCommand creates a global application slash command.
func RouteCreateGlobalCommand(appID discord.Snowflake) *Route {
	return &Route{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("/applications/%s/commands", appID),
		BucketKey:  "/applications/:app_id/commands",
		MajorParam: appID.String(),
	}
}

// RouteCreateGuildCommand creates a guild-specific application slash command.
func RouteCreateGuildCommand(appID, guildID discord.Snowflake) *Route {
	return &Route{
		Method:     http.MethodPost,
		Path:       fmt.Sprintf("/applications/%s/guilds/%s/commands", appID, guildID),
		BucketKey:  fmt.Sprintf("/applications/:app_id/guilds/%s/commands", guildID),
		MajorParam: guildID.String(),
	}
}

// RouteGetGatewayBot returns the gateway websocket URL and recommended shard count.
func RouteGetGatewayBot() *Route {
	return &Route{
		Method:     http.MethodGet,
		Path:       "/gateway/bot",
		BucketKey:  "/gateway/bot",
		MajorParam: "global",
	}
}
