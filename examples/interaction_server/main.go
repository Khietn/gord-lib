package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Khietn/gord-lib"
	"github.com/Khietn/gord-lib/discord"
)

func main() {
	// Automatically load .env file
	if err := gord.LoadEnv(); err != nil {
		fmt.Printf("Notice: .env file not loaded (%v), using system environment\n", err)
	}

	appIDStr := os.Getenv("DISCORD_APP_ID")
	if appIDStr == "" {
		appIDStr = "888070652857286667"
	}
	appSnowflake, err := discord.ParseSnowflake(appIDStr)
	if err != nil {
		log.Fatalf("Invalid app snowflake: %v", err)
	}

	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")
	if publicKey == "" {
		publicKey = "1e8fa8b5ba9c6b38d6e2769fe2d21b84b6b11ed2e4cbdb59bf37f02c59991a9c"
	}

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		token = "OPTIONAL_FOR_HTTP_ONLY_INTERACTIONS"
	}

	// Create client with Application ID and Public Key
	client, err := gord.New(token,
		gord.WithApplicationID(appSnowflake),
		gord.WithPublicKey(publicKey),
	)
	if err != nil {
		log.Fatalf("Failed to initialize gord client: %v", err)
	}

	// Define interaction handler for HTTP Interactions (Slash commands & PING)
	handler, err := client.HTTPInteractionHandler(func(interaction *discord.Interaction) *discord.InteractionResponse {
		fmt.Printf("Received interaction ID: %s, Type: %d\n", interaction.ID, interaction.Type)

		// Build a rich embed response
		embed, err := discord.NewEmbedBuilder().
			SetTitle("Pong! 🏓").
			SetDescription("Hello from gord-lib Interaction Webhook Server!").
			SetColor(0x5865F2).
			AddField("Latency", "HTTP Only (Zero WebSocket overhead)", true).
			Build()
		if err != nil {
			fmt.Printf("Error building embed: %v\n", err)
		}

		return &discord.InteractionResponse{
			Type: discord.InteractionCallbackTypeChannelMessageWithSource,
			Data: &discord.InteractionCallbackData{
				Content: "🏓 Pong!",
				Embeds:  []discord.Embed{embed},
			},
		}
	})
	if err != nil {
		log.Fatalf("Failed to create interaction handler: %v", err)
	}

	// Register HTTP endpoint for Discord Interactions URL
	http.HandleFunc("/interactions", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Interaction Webhook Server running on :%s/interactions\n", port)
	fmt.Println("Configure your Discord Developer Portal Interactions Endpoint URL to:")
	fmt.Printf("👉 https://<your-public-tunnel-domain>/interactions\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
