package events

import (
	"github.com/Khietn/gord-lib/discord"
)

// Event is the common interface implemented by all Gateway dispatched events.
type Event interface {
	EventName() string
}

// Ready is dispatched when the client completes the initial handshake.
type Ready struct {
	User             discord.User      `json:"user"`
	SessionID        string            `json:"session_id"`
	ResumeGatewayURL string            `json:"resume_gateway_url"`
	Shard            *[2]int           `json:"shard,omitempty"`
	ApplicationID    discord.Snowflake `json:"application_id"`
}

func (Ready) EventName() string { return "READY" }

// MessageCreate is dispatched when a message is created.
type MessageCreate struct {
	discord.Message
}

func (MessageCreate) EventName() string { return "MESSAGE_CREATE" }

// MessageUpdate is dispatched when a message is edited.
type MessageUpdate struct {
	discord.Message
}

func (MessageUpdate) EventName() string { return "MESSAGE_UPDATE" }

// MessageDelete is dispatched when a message is deleted.
type MessageDelete struct {
	ID        discord.Snowflake  `json:"id"`
	ChannelID discord.Snowflake  `json:"channel_id"`
	GuildID   *discord.Snowflake `json:"guild_id,omitempty"`
}

func (MessageDelete) EventName() string { return "MESSAGE_DELETE" }

// InteractionCreate is dispatched when a user triggers a slash command or component.
type InteractionCreate struct {
	discord.Interaction
}

func (InteractionCreate) EventName() string { return "INTERACTION_CREATE" }

// GuildCreate is dispatched when a guild becomes available.
type GuildCreate struct {
	discord.Guild
}

func (GuildCreate) EventName() string { return "GUILD_CREATE" }

// GuildDelete is dispatched when a guild becomes unavailable or the bot leaves.
type GuildDelete struct {
	ID          discord.Snowflake `json:"id"`
	Unavailable bool              `json:"unavailable"`
}

func (GuildDelete) EventName() string { return "GUILD_DELETE" }
