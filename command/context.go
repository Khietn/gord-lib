package command

import (
	"context"

	"github.com/Khietn/gord-lib"
	"github.com/Khietn/gord-lib/discord"
	"github.com/Khietn/gord-lib/events"
)

// Context encapsulates invocation data, providing a unified facade for both
// prefix-based message commands and Discord Slash Command interactions.
type Context struct {
	Client      *gord.Client
	Message     *events.MessageCreate
	Interaction *events.InteractionCreate
	Args        []string
	subcommand  string
}

// NewMessageContext constructs a Context from a MessageCreate event.
func NewMessageContext(client *gord.Client, msg *events.MessageCreate, args []string) *Context {
	var sub string
	if len(args) > 0 {
		sub = args[0]
	}
	return &Context{
		Client:     client,
		Message:    msg,
		Args:       args,
		subcommand: sub,
	}
}

// ChannelID returns the ID of the channel where the command was issued.
func (c *Context) ChannelID() discord.Snowflake {
	if c.Message != nil {
		return c.Message.ChannelID
	}
	if c.Interaction != nil && c.Interaction.ChannelID != nil {
		return *c.Interaction.ChannelID
	}
	return 0
}

// Author returns the user who issued the command.
func (c *Context) Author() discord.User {
	if c.Message != nil {
		return c.Message.Author
	}
	if c.Interaction != nil && c.Interaction.User != nil {
		return *c.Interaction.User
	}
	return discord.User{}
}

// SubcommandName returns the first argument as the subcommand identifier.
func (c *Context) SubcommandName() string {
	return c.subcommand
}

// Reply sends a reply message to the author in the current channel.
func (c *Context) Reply(ctx context.Context, content string) (*discord.Message, error) {
	if c.Client != nil && c.Client.Rest != nil && c.Message != nil {
		payload := discord.NewMessageBuilder().
			SetContent(content).
			ReplyTo(c.Message.ID, c.Message.ChannelID).
			Build()
		return c.Client.Rest.CreateMessage(ctx, c.Message.ChannelID, payload)
	}

	return nil, nil
}

// ReplyEmbed sends a rich embed response to the author in the current channel.
func (c *Context) ReplyEmbed(ctx context.Context, embed discord.Embed) (*discord.Message, error) {
	if c.Client != nil && c.Client.Rest != nil && c.Message != nil {
		payload := discord.NewMessageBuilder().
			AddEmbed(embed).
			ReplyTo(c.Message.ID, c.Message.ChannelID).
			Build()
		return c.Client.Rest.CreateMessage(ctx, c.Message.ChannelID, payload)
	}

	return nil, nil
}
