package gord

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Khietn/gord-lib/discord"
)

func TestNew_ClientOptions(t *testing.T) {
	// Empty token error
	_, err := New("")
	if err == nil {
		t.Errorf("expected error with empty token, got nil")
	}

	appID := discord.Snowflake(123456789)
	pubKey := "1e8fa8b5ba9c6b38d6e2769fe2d21b84b6b11ed2e4cbdb59bf37f02c59991a9c"

	client, err := New("valid-token",
		WithApplicationID(appID),
		WithPublicKey(pubKey),
		WithIntents(discord.IntentsDefault|discord.IntentMessageContent),
		WithWorkerCount(4),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	if client.applicationID != appID {
		t.Errorf("expected application ID %v, got %v", appID, client.applicationID)
	}
	if client.publicKey != pubKey {
		t.Errorf("expected public key %s, got %s", pubKey, client.publicKey)
	}
	if client.Rest == nil {
		t.Errorf("expected initialized REST client")
	}
	if client.Gateway == nil {
		t.Errorf("expected initialized Gateway")
	}
	if client.Dispatcher == nil {
		t.Errorf("expected initialized Dispatcher")
	}

	// Test HTTPInteractionHandler without public key
	emptyClient, _ := New("test")
	_, err = emptyClient.HTTPInteractionHandler(nil)
	if err == nil {
		t.Errorf("expected error when public key is missing")
	}

	// Test HTTPInteractionHandler with public key
	handler, err := client.HTTPInteractionHandler(nil)
	if err != nil || handler == nil {
		t.Errorf("expected valid HTTP handler, got: %v", err)
	}
}

func TestLoadEnv(t *testing.T) {
	// Create temporary .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.test")

	content := `
# Comment line
TEST_KEY_ONE=value1
TEST_KEY_TWO="quoted_value"
TEST_KEY_THREE='single_quoted'
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp .env: %v", err)
	}

	if err := LoadEnv(envPath); err != nil {
		t.Fatalf("LoadEnv failed: %v", err)
	}

	if os.Getenv("TEST_KEY_ONE") != "value1" {
		t.Errorf("expected TEST_KEY_ONE=value1, got %q", os.Getenv("TEST_KEY_ONE"))
	}
	if os.Getenv("TEST_KEY_TWO") != "quoted_value" {
		t.Errorf("expected TEST_KEY_TWO=quoted_value, got %q", os.Getenv("TEST_KEY_TWO"))
	}
	if os.Getenv("TEST_KEY_THREE") != "single_quoted" {
		t.Errorf("expected TEST_KEY_THREE=single_quoted, got %q", os.Getenv("TEST_KEY_THREE"))
	}

	// Non-existent file should error
	if err := LoadEnv("non-existent-file-path-12345.env"); err == nil {
		t.Errorf("expected error on non-existent file")
	}
}
