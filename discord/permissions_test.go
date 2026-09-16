package discord

import (
	"encoding/json"
	"testing"
)

func TestPermissions_Bitwise(t *testing.T) {
	perms := PermissionSendMessages.Add(PermissionViewChannel, PermissionAttachFiles)

	if !perms.Has(PermissionSendMessages) {
		t.Errorf("expected perms to have PermissionSendMessages")
	}
	if !perms.Has(PermissionViewChannel) {
		t.Errorf("expected perms to have PermissionViewChannel")
	}
	if perms.Has(PermissionAdministrator) {
		t.Errorf("perms should not have PermissionAdministrator")
	}

	// Remove permission
	perms = perms.Remove(PermissionSendMessages)
	if perms.Has(PermissionSendMessages) {
		t.Errorf("expected perms to NOT have PermissionSendMessages after removal")
	}

	// Administrator override
	adminPerms := PermissionAdministrator
	if !adminPerms.Has(PermissionBanMembers) {
		t.Errorf("administrator should have all permissions")
	}
}

func TestPermissions_JSON(t *testing.T) {
	perms := PermissionSendMessages | PermissionViewChannel
	data, err := json.Marshal(perms)
	if err != nil {
		t.Fatalf("failed to marshal perms: %v", err)
	}

	var parsed Permissions
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("failed to unmarshal perms: %v", err)
	}

	if parsed != perms {
		t.Fatalf("expected %d, got %d", perms, parsed)
	}
}
