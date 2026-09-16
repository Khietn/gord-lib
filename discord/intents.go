package discord

// GatewayIntents represents bitwise integer flags for Discord Gateway Intents.
type GatewayIntents uint64

// Gateway Intent constants.
const (
	IntentGuilds                      GatewayIntents = 1 << 0
	IntentGuildMembers                GatewayIntents = 1 << 1 // Privileged
	IntentGuildModeration             GatewayIntents = 1 << 2
	IntentGuildEmojisAndStickers      GatewayIntents = 1 << 3
	IntentGuildIntegrations           GatewayIntents = 1 << 4
	IntentGuildWebhooks               GatewayIntents = 1 << 5
	IntentGuildInvites                GatewayIntents = 1 << 6
	IntentGuildVoiceStates            GatewayIntents = 1 << 7
	IntentGuildPresences              GatewayIntents = 1 << 8 // Privileged
	IntentGuildMessages               GatewayIntents = 1 << 9
	IntentGuildMessageReactions       GatewayIntents = 1 << 10
	IntentGuildMessageTyping          GatewayIntents = 1 << 11
	IntentDirectMessages              GatewayIntents = 1 << 12
	IntentDirectMessageReactions      GatewayIntents = 1 << 13
	IntentDirectMessageTyping         GatewayIntents = 1 << 14
	IntentMessageContent              GatewayIntents = 1 << 15 // Privileged
	IntentGuildScheduledEvents        GatewayIntents = 1 << 16
	IntentAutoModerationConfiguration GatewayIntents = 1 << 20
	IntentAutoModerationExecution     GatewayIntents = 1 << 21
	IntentGuildMessagePolls           GatewayIntents = 1 << 24
	IntentDirectMessagePolls          GatewayIntents = 1 << 25

	// IntentsPrivileged includes all intents requiring approval on the Discord Developer Portal.
	IntentsPrivileged = IntentGuildMembers | IntentGuildPresences | IntentMessageContent

	// IntentsDefault includes all non-privileged intents.
	IntentsDefault = IntentGuilds | IntentGuildModeration | IntentGuildEmojisAndStickers |
		IntentGuildIntegrations | IntentGuildWebhooks | IntentGuildInvites |
		IntentGuildVoiceStates | IntentGuildMessages | IntentGuildMessageReactions |
		IntentGuildMessageTyping | IntentDirectMessages | IntentDirectMessageReactions |
		IntentDirectMessageTyping | IntentGuildScheduledEvents |
		IntentAutoModerationConfiguration | IntentAutoModerationExecution |
		IntentGuildMessagePolls | IntentDirectMessagePolls

	// IntentsAll includes every single intent.
	IntentsAll = IntentsDefault | IntentsPrivileged
)

// Has checks whether i contains the specified target intents.
func (i GatewayIntents) Has(target GatewayIntents) bool {
	return (i & target) == target
}

// Add adds one or more intents to i and returns the updated bitmask.
func (i GatewayIntents) Add(intents ...GatewayIntents) GatewayIntents {
	res := i
	for _, intent := range intents {
		res |= intent
	}
	return res
}

// Remove removes one or more intents from i and returns the updated bitmask.
func (i GatewayIntents) Remove(intents ...GatewayIntents) GatewayIntents {
	res := i
	for _, intent := range intents {
		res &= ^intent
	}
	return res
}
