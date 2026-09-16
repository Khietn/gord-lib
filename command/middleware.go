package command

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Khietn/gord-lib/discord"
)

// HandlerFunc defines the execution signature for a command handler.
type HandlerFunc func(ctx context.Context, cmdCtx *Context) error

// Middleware intercepts command execution in the Chain of Responsibility pattern.
type Middleware func(next HandlerFunc) HandlerFunc

// LoggingMiddleware logs execution details and latency for every command.
func LoggingMiddleware(logFn func(format string, args ...any)) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, cmdCtx *Context) (err error) {
			start := time.Now()
			author := cmdCtx.Author()
			cmdName := "unknown"
			if len(cmdCtx.Args) > 0 {
				cmdName = cmdCtx.Args[0]
			}

			defer func() {
				duration := time.Since(start)
				if err != nil {
					logFn("[gord/command] command=%s user=%s (id=%s) duration=%v error=%v",
						cmdName, author.Username, author.ID, duration, err)
				} else {
					logFn("[gord/command] command=%s user=%s (id=%s) duration=%v status=success",
						cmdName, author.Username, author.ID, duration)
				}
			}()

			return next(ctx, cmdCtx)
		}
	}
}

// RecoveryMiddleware catches any panics inside command handlers and converts them to errors.
func RecoveryMiddleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, cmdCtx *Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic recovered during command execution: %v", r)
				}
			}()
			return next(ctx, cmdCtx)
		}
	}
}

// RequirePermissions enforces that the author possesses the required Discord bitwise permissions.
func RequirePermissions(required discord.Permissions, getMemberPerms func(cmdCtx *Context) discord.Permissions) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, cmdCtx *Context) error {
			if getMemberPerms == nil {
				return next(ctx, cmdCtx)
			}

			perms := getMemberPerms(cmdCtx)
			if !perms.Has(required) {
				return fmt.Errorf("permission denied: you are missing required permissions (%d)", perms.Missing(required))
			}
			return next(ctx, cmdCtx)
		}
	}
}

// CooldownTracker maintains per-user rate limiting cooldowns for commands.
type CooldownTracker struct {
	mu        sync.Mutex
	cooldowns map[discord.Snowflake]time.Time
	duration  time.Duration
}

// NewCooldownTracker constructs a new CooldownTracker with the specified window.
func NewCooldownTracker(duration time.Duration) *CooldownTracker {
	return &CooldownTracker{
		cooldowns: make(map[discord.Snowflake]time.Time),
		duration:  duration,
	}
}

// Middleware creates a cooldown interceptor that limits invocations per author.
func (c *CooldownTracker) Middleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, cmdCtx *Context) error {
			author := cmdCtx.Author()
			if !author.ID.IsValid() {
				return next(ctx, cmdCtx)
			}

			c.mu.Lock()
			expiry, active := c.cooldowns[author.ID]
			now := time.Now()

			if active && now.Before(expiry) {
				remaining := time.Until(expiry)
				c.mu.Unlock()
				return fmt.Errorf("command is on cooldown: please wait %.1f seconds", remaining.Seconds())
			}

			c.cooldowns[author.ID] = now.Add(c.duration)
			c.mu.Unlock()

			return next(ctx, cmdCtx)
		}
	}
}
