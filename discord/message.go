package discord

import (
	"time"
)

// MessageFlags represents bitwise flags on a Discord Message.
type MessageFlags uint64

const (
	MessageFlagCrossposted           MessageFlags = 1 << 0
	MessageFlagIsCrosspost           MessageFlags = 1 << 1
	MessageFlagSuppressEmbeds        MessageFlags = 1 << 2
	MessageFlagSourceMessageDeleted  MessageFlags = 1 << 3
	MessageFlagUrgent                MessageFlags = 1 << 4
	MessageFlagHasThread             MessageFlags = 1 << 5
	MessageFlagEphemeral             MessageFlags = 1 << 6
	MessageFlagLoading               MessageFlags = 1 << 7
	MessageFlagFailedToMentionRoles  MessageFlags = 1 << 8
	MessageFlagSuppressNotifications MessageFlags = 1 << 12
	MessageFlagIsVoiceMessage        MessageFlags = 1 << 13
)

// User represents a Discord user account.
type User struct {
	ID            Snowflake `json:"id"`
	Username      string    `json:"username"`
	Discriminator string    `json:"discriminator"`
	GlobalName    *string   `json:"global_name,omitempty"`
	Avatar        *string   `json:"avatar,omitempty"`
	Bot           bool      `json:"bot,omitempty"`
	System        bool      `json:"system,omitempty"`
	MFAEnabled    bool      `json:"mfa_enabled,omitempty"`
	Banner        *string   `json:"banner,omitempty"`
	AccentColor   *int      `json:"accent_color,omitempty"`
	Locale        string    `json:"locale,omitempty"`
	Verified      bool      `json:"verified,omitempty"`
	Flags         int       `json:"flags,omitempty"`
	PublicFlags   int       `json:"public_flags,omitempty"`
}

// Channel represents a Discord guild or DM channel.
type Channel struct {
	ID               Snowflake  `json:"id"`
	Type             int        `json:"type"`
	GuildID          *Snowflake `json:"guild_id,omitempty"`
	Position         *int       `json:"position,omitempty"`
	Name             string     `json:"name,omitempty"`
	Topic            *string    `json:"topic,omitempty"`
	NSFW             bool       `json:"nsfw,omitempty"`
	LastMessageID    *Snowflake `json:"last_message_id,omitempty"`
	Bitrate          *int       `json:"bitrate,omitempty"`
	UserLimit        *int       `json:"user_limit,omitempty"`
	RateLimitPerUser *int       `json:"rate_limit_per_user,omitempty"`
	ParentID         *Snowflake `json:"parent_id,omitempty"`
	OwnerID          *Snowflake `json:"owner_id,omitempty"`
}

// Guild represents a Discord guild (server).
type Guild struct {
	ID                       Snowflake   `json:"id"`
	Name                     string      `json:"name"`
	Icon                     *string     `json:"icon,omitempty"`
	OwnerID                  Snowflake   `json:"owner_id"`
	Permissions              Permissions `json:"permissions,omitempty"`
	AFKChannelID             *Snowflake  `json:"afk_channel_id,omitempty"`
	AFKTimeout               int         `json:"afk_timeout"`
	MemberCount              int         `json:"member_count,omitempty"`
	Description              *string     `json:"description,omitempty"`
	PremiumTier              int         `json:"premium_tier"`
	PreferredLocale          string      `json:"preferred_locale"`
	ApproximateMemberCount   int         `json:"approximate_member_count,omitempty"`
	ApproximatePresenceCount int         `json:"approximate_presence_count,omitempty"`
}

// Message represents a message sent in a channel within Discord.
type Message struct {
	ID                Snowflake    `json:"id"`
	ChannelID         Snowflake    `json:"channel_id"`
	GuildID           *Snowflake   `json:"guild_id,omitempty"`
	Author            User         `json:"author"`
	Content           string       `json:"content"`
	Timestamp         time.Time    `json:"timestamp"`
	EditedTimestamp   *time.Time   `json:"edited_timestamp,omitempty"`
	TTS               bool         `json:"tts"`
	MentionEveryone   bool         `json:"mention_everyone"`
	Mentions          []User       `json:"mentions"`
	Embeds            []Embed      `json:"embeds"`
	Pinned            bool         `json:"pinned"`
	WebhookID         *Snowflake   `json:"webhook_id,omitempty"`
	Type              int          `json:"type"`
	Flags             MessageFlags `json:"flags,omitempty"`
	ReferencedMessage *Message     `json:"referenced_message,omitempty"`
}

// MessageCreatePayload represents the payload sent to create a message in a channel.
type MessageCreatePayload struct {
	Content          string       `json:"content,omitempty"`
	TTS              bool         `json:"tts,omitempty"`
	Embeds           []Embed      `json:"embeds,omitempty"`
	Flags            MessageFlags `json:"flags,omitempty"`
	MessageReference *MessageRef  `json:"message_reference,omitempty"`
}

// MessageRef represents a reference to another message.
type MessageRef struct {
	MessageID Snowflake `json:"message_id,omitempty"`
	ChannelID Snowflake `json:"channel_id,omitempty"`
	GuildID   Snowflake `json:"guild_id,omitempty"`
}

// MessageBuilder builds a MessageCreatePayload using fluent chaining.
type MessageBuilder struct {
	payload MessageCreatePayload
}

// NewMessageBuilder creates a new message builder.
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{}
}

// SetContent sets the raw text content of the message.
func (b *MessageBuilder) SetContent(content string) *MessageBuilder {
	b.payload.Content = content
	return b
}

// SetTTS sets the text-to-speech flag.
func (b *MessageBuilder) SetTTS(tts bool) *MessageBuilder {
	b.payload.TTS = tts
	return b
}

// AddEmbed adds an Embed to the message.
func (b *MessageBuilder) AddEmbed(embed Embed) *MessageBuilder {
	b.payload.Embeds = append(b.payload.Embeds, embed)
	return b
}

// SetFlags sets the message flags (e.g. MessageFlagSuppressNotifications).
func (b *MessageBuilder) SetFlags(flags MessageFlags) *MessageBuilder {
	b.payload.Flags = flags
	return b
}

// ReplyTo sets the message reference to reply to an existing message.
func (b *MessageBuilder) ReplyTo(msgID, channelID Snowflake) *MessageBuilder {
	b.payload.MessageReference = &MessageRef{
		MessageID: msgID,
		ChannelID: channelID,
	}
	return b
}

// Build returns the constructed MessageCreatePayload.
func (b *MessageBuilder) Build() MessageCreatePayload {
	return b.payload
}
