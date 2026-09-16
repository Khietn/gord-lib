package discord

// InteractionType represents the type of interaction.
type InteractionType int

const (
	InteractionTypePing                           InteractionType = 1
	InteractionTypeApplicationCommand             InteractionType = 2
	InteractionTypeMessageComponent               InteractionType = 3
	InteractionTypeApplicationCommandAutocomplete InteractionType = 4
	InteractionTypeModalSubmit                    InteractionType = 5
)

// InteractionCallbackType represents the type of response to an interaction.
type InteractionCallbackType int

const (
	InteractionCallbackTypePong                                 InteractionCallbackType = 1
	InteractionCallbackTypeChannelMessageWithSource             InteractionCallbackType = 4
	InteractionCallbackTypeDeferredChannelMessageWithSource     InteractionCallbackType = 5
	InteractionCallbackTypeDeferredUpdateMessage                InteractionCallbackType = 6
	InteractionCallbackTypeUpdateMessage                        InteractionCallbackType = 7
	InteractionCallbackTypeApplicationCommandAutocompleteResult InteractionCallbackType = 8
	InteractionCallbackTypeModal                                InteractionCallbackType = 9
)

// ComponentType represents a UI component type.
type ComponentType int

const (
	ComponentTypeActionRow         ComponentType = 1
	ComponentTypeButton            ComponentType = 2
	ComponentTypeStringSelect      ComponentType = 3
	ComponentTypeTextInput         ComponentType = 4
	ComponentTypeUserSelect        ComponentType = 5
	ComponentTypeRoleSelect        ComponentType = 6
	ComponentTypeMentionableSelect ComponentType = 7
	ComponentTypeChannelSelect     ComponentType = 8
)

// ButtonStyle represents the color style of a button component.
type ButtonStyle int

const (
	ButtonStylePrimary   ButtonStyle = 1 // Blurple
	ButtonStyleSecondary ButtonStyle = 2 // Grey
	ButtonStyleSuccess   ButtonStyle = 3 // Green
	ButtonStyleDanger    ButtonStyle = 4 // Red
	ButtonStyleLink      ButtonStyle = 5 // URL
)

// Component represents an interactive UI component (Button, Select, ActionRow).
type Component struct {
	Type       ComponentType `json:"type"`
	CustomID   string        `json:"custom_id,omitempty"`
	Disabled   bool          `json:"disabled,omitempty"`
	Style      ButtonStyle   `json:"style,omitempty"`
	Label      string        `json:"label,omitempty"`
	URL        string        `json:"url,omitempty"`
	Components []Component   `json:"components,omitempty"`
}

// Interaction represents an incoming Discord interaction (Slash command, button click, etc.).
type Interaction struct {
	ID            Snowflake       `json:"id"`
	ApplicationID Snowflake       `json:"application_id"`
	Type          InteractionType `json:"type"`
	Data          any             `json:"data,omitempty"`
	GuildID       *Snowflake      `json:"guild_id,omitempty"`
	ChannelID     *Snowflake      `json:"channel_id,omitempty"`
	Token         string          `json:"token"`
	Version       int             `json:"version"`
	Message       *Message        `json:"message,omitempty"`
	User          *User           `json:"user,omitempty"`
}

// InteractionResponse represents the payload sent back to acknowledge/respond to an interaction.
type InteractionResponse struct {
	Type InteractionCallbackType  `json:"type"`
	Data *InteractionCallbackData `json:"data,omitempty"`
}

// InteractionCallbackData represents the response data for an interaction.
type InteractionCallbackData struct {
	TTS        bool         `json:"tts,omitempty"`
	Content    string       `json:"content,omitempty"`
	Embeds     []Embed      `json:"embeds,omitempty"`
	Flags      MessageFlags `json:"flags,omitempty"`
	Components []Component  `json:"components,omitempty"`
}
