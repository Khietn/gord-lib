package discord

import (
	"bytes"
	"fmt"
	"strconv"
)

// Permissions represents a bitwise integer representing Discord permission flags.
type Permissions uint64

// Discord Permissions bit flags.
const (
	PermissionCreateInstantInvite              Permissions = 1 << 0
	PermissionKickMembers                      Permissions = 1 << 1
	PermissionBanMembers                       Permissions = 1 << 2
	PermissionAdministrator                    Permissions = 1 << 3
	PermissionManageChannels                   Permissions = 1 << 4
	PermissionManageGuild                      Permissions = 1 << 5
	PermissionAddReactions                     Permissions = 1 << 6
	PermissionViewAuditLog                     Permissions = 1 << 7
	PermissionPrioritySpeaker                  Permissions = 1 << 8
	PermissionStream                           Permissions = 1 << 9
	PermissionViewChannel                      Permissions = 1 << 10
	PermissionSendMessages                     Permissions = 1 << 11
	PermissionSendTTSMessages                  Permissions = 1 << 12
	PermissionManageMessages                   Permissions = 1 << 13
	PermissionEmbedLinks                       Permissions = 1 << 14
	PermissionAttachFiles                      Permissions = 1 << 15
	PermissionReadMessageHistory               Permissions = 1 << 16
	PermissionMentionEveryone                  Permissions = 1 << 17
	PermissionUseExternalEmojis                Permissions = 1 << 18
	PermissionViewGuildInsights                Permissions = 1 << 19
	PermissionConnect                          Permissions = 1 << 20
	PermissionSpeak                            Permissions = 1 << 21
	PermissionMuteMembers                      Permissions = 1 << 22
	PermissionDeafenMembers                    Permissions = 1 << 23
	PermissionMoveMembers                      Permissions = 1 << 24
	PermissionUseVAD                           Permissions = 1 << 25
	PermissionChangeNickname                   Permissions = 1 << 26
	PermissionManageNicknames                  Permissions = 1 << 27
	PermissionManageRoles                      Permissions = 1 << 28
	PermissionManageWebhooks                   Permissions = 1 << 29
	PermissionManageGuildExpressions           Permissions = 1 << 30
	PermissionUseApplicationCommands           Permissions = 1 << 31
	PermissionRequestToSpeak                   Permissions = 1 << 32
	PermissionManageEvents                     Permissions = 1 << 33
	PermissionManageThreads                    Permissions = 1 << 34
	PermissionCreatePublicThreads              Permissions = 1 << 35
	PermissionCreatePrivateThreads             Permissions = 1 << 36
	PermissionUseExternalStickers              Permissions = 1 << 37
	PermissionSendMessagesInThreads            Permissions = 1 << 38
	PermissionUseEmbeddedActivities            Permissions = 1 << 39
	PermissionModerateMembers                  Permissions = 1 << 40
	PermissionViewCreatorMonetizationAnalytics Permissions = 1 << 41
	PermissionUseSoundboard                    Permissions = 1 << 42
	PermissionUseExternalSounds                Permissions = 1 << 45
	PermissionSendVoiceMessages                Permissions = 1 << 46
	PermissionSendPolls                        Permissions = 1 << 49
)

// Has checks whether p contains the target permission.
// If p has PermissionAdministrator, it returns true for any checked permission.
func (p Permissions) Has(target Permissions) bool {
	if p&PermissionAdministrator == PermissionAdministrator {
		return true
	}
	return (p & target) == target
}

// Add adds one or more permissions and returns the updated bitset.
func (p Permissions) Add(perms ...Permissions) Permissions {
	res := p
	for _, perm := range perms {
		res |= perm
	}
	return res
}

// Remove removes one or more permissions and returns the updated bitset.
func (p Permissions) Remove(perms ...Permissions) Permissions {
	res := p
	for _, perm := range perms {
		res &= ^perm
	}
	return res
}

// Missing returns the subset of required permissions that are not present in p.
func (p Permissions) Missing(required Permissions) Permissions {
	if p.Has(PermissionAdministrator) {
		return 0
	}
	return required &^ p
}

// MarshalJSON marshals Permissions as a string (Discord API standard).
func (p Permissions) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(strconv.FormatUint(uint64(p), 10))), nil
}

// UnmarshalJSON unmarshals Permissions from a string or uint64.
func (p *Permissions) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*p = 0
		return nil
	}

	if data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}

	v, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to unmarshal permissions %q: %w", string(data), err)
	}

	*p = Permissions(v)
	return nil
}
