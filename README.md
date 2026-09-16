# gord-lib

A modern, high-performance, and idiomatic Discord API library for Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/Khietn/gord-lib.svg)](https://pkg.go.dev/github.com/Khietn/gord-lib)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Core Architectural Features & Design Patterns

- **Command Pattern (`command`)**: First-class command router with isolated, modular command handlers.
- **Chain of Responsibility (Middlewares)**: Interceptor pipeline for panic recovery, latency telemetry, permissions enforcement, and per-user cooldowns.
- **Composite Pattern (Subcommands)**: Nest subcommands cleanly into groups (`!config get`, `!config set`).
- **REST Client & Rate Limiter (`rest`)**: Route-aware per-bucket rate limiting (`X-RateLimit-Bucket`) compliant with Discord v10.
- **Gateway WebSocket (`gateway`)**: Finite State Machine (FSM), heartbeat jitter, and zombie connection detection with Gorilla WebSocket.
- **Worker Pool Dispatcher (`events`)**: Concurrent, non-blocking event dispatching with Go generics (`Register[E Event]`).
- **Fluent Builders (`discord`)**: Type-safe builders for Messages and Embeds with boundary validation.
- **Ed25519 Security**: Cryptographic interaction signature verification for zero-overhead HTTP Webhook bots.

## Installation

```bash
go get github.com/Khietn/gord-lib
```

## Quick Start (Command Pattern Bot)

```go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Khietn/gord-lib"
	"github.com/Khietn/gord-lib/command"
	"github.com/Khietn/gord-lib/discord"
	"github.com/Khietn/gord-lib/events"
)

type PingCommand struct{}

func (p *PingCommand) Name() string        { return "ping" }
func (p *PingCommand) Description() string { return "Replies with Pong!" }
func (p *PingCommand) Execute(ctx context.Context, cmdCtx *command.Context) error {
	_, err := cmdCtx.Reply(ctx, "Pong! 🏓")
	return err
}

func main() {
	_ = gord.LoadEnv()
	client, err := gord.New(os.Getenv("DISCORD_BOT_TOKEN"),
		gord.WithIntents(discord.IntentsDefault|discord.IntentMessageContent),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Command Router (Invoker) + Middlewares (Chain of Responsibility)
	router := command.NewRouter("/")
	router.Use(command.RecoveryMiddleware())
	router.Use(command.NewCooldownTracker(2 * time.Second).Middleware())
	router.Register(&PingCommand{})

	gord.On(client, func(ctx context.Context, e events.MessageCreate) {
		router.HandleMessage(ctx, client, e)
	})

	if err := client.Open(context.Background()); err != nil {
		log.Fatal(err)
	}

	select {}
}
```

## Running the Examples

### 1. Gateway Bot with Diagnostic Commands
```bash
go run ./examples/ping_bot/main.go
```
Available commands: `/help`, `/ping`, `/embed`, `/user`, `/stats`, `/config get`, `/testall`.

### 2. HTTP Interaction Webhook Server
```bash
go run ./examples/interaction_server/main.go
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
