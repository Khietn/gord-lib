package command

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Khietn/gord-lib/discord"
	"github.com/Khietn/gord-lib/events"
)

// Mock command
type mockCmd struct {
	name        string
	description string
	executed    bool
	panics      bool
}

func (m *mockCmd) Name() string        { return m.name }
func (m *mockCmd) Description() string { return m.description }
func (m *mockCmd) Execute(ctx context.Context, cmdCtx *Context) error {
	if m.panics {
		panic("simulated command panic")
	}
	m.executed = true
	return nil
}

func TestRouter_Dispatch(t *testing.T) {
	router := NewRouter() // Defaults to "/"
	if router.Prefix() != "/" {
		t.Errorf("expected default prefix '/', got %s", router.Prefix())
	}

	cmd := &mockCmd{name: "ping", description: "test ping"}
	router.Register(cmd)

	ctx := context.Background()

	// Non-matching message
	router.HandleMessage(ctx, nil, events.MessageCreate{
		Message: discord.Message{Content: "hello world"},
	})
	if cmd.executed {
		t.Errorf("command should not execute for non-matching message")
	}

	// Message with wrong prefix
	router.HandleMessage(ctx, nil, events.MessageCreate{
		Message: discord.Message{Content: "!ping"},
	})
	if cmd.executed {
		t.Errorf("command should not execute with wrong prefix !")
	}

	// Matching command with /
	router.HandleMessage(ctx, nil, events.MessageCreate{
		Message: discord.Message{Content: "/ping"},
	})
	if !cmd.executed {
		t.Errorf("expected command /ping to execute")
	}

	// Commands list
	cmds := router.Commands()
	if len(cmds) != 1 || cmds[0].Name() != "ping" {
		t.Errorf("expected 1 command in router")
	}
}

func TestComposite_GroupSubcommands(t *testing.T) {
	group := NewGroup("config", "configuration group")
	subGet := &mockCmd{name: "get", description: "get config"}
	subSet := &mockCmd{name: "set", description: "set config"}

	group.RegisterSubcommands(subGet, subSet)

	router := NewRouter("!")
	router.Register(group)

	ctx := context.Background()

	// Test missing subcommand
	err := group.Execute(ctx, &Context{Args: []string{}})
	if err == nil || !strings.Contains(err.Error(), "missing subcommand") {
		t.Errorf("expected missing subcommand error, got %v", err)
	}

	// Test unknown subcommand
	err = group.Execute(ctx, &Context{Args: []string{"invalid"}})
	if err == nil || !strings.Contains(err.Error(), "unknown subcommand") {
		t.Errorf("expected unknown subcommand error, got %v", err)
	}

	// Test executing valid subcommand
	err = group.Execute(ctx, &Context{Args: []string{"get"}})
	if err != nil {
		t.Errorf("expected subcommand 'get' to execute without error: %v", err)
	}
	if !subGet.executed {
		t.Errorf("expected subGet to be executed")
	}
}

func TestMiddleware_ChainAndRecovery(t *testing.T) {
	router := NewRouter()
	router.Use(RecoveryMiddleware())

	loggedMsg := ""
	router.Use(LoggingMiddleware(func(format string, args ...any) {
		loggedMsg = "logged"
	}))

	panickingCmd := &mockCmd{name: "boom", panics: true}
	router.Register(panickingCmd)

	ctx := context.Background()

	// Executing panicking command should not crash due to RecoveryMiddleware
	router.HandleMessage(ctx, nil, events.MessageCreate{
		Message: discord.Message{Content: "/boom"},
	})

	if loggedMsg != "logged" {
		t.Errorf("expected LoggingMiddleware to execute")
	}
}

func TestMiddleware_Cooldown(t *testing.T) {
	cooldown := NewCooldownTracker(100 * time.Millisecond)

	var runs int
	handler := cooldown.Middleware()(func(ctx context.Context, cmdCtx *Context) error {
		runs++
		return nil
	})

	ctx := context.Background()
	cmdCtx := &Context{
		Message: &events.MessageCreate{
			Message: discord.Message{
				Author: discord.User{ID: discord.Snowflake(12345)},
			},
		},
	}

	// First call succeeds
	if err := handler(ctx, cmdCtx); err != nil {
		t.Fatalf("first call should succeed: %v", err)
	}

	// Second immediate call fails with cooldown error
	err := handler(ctx, cmdCtx)
	if err == nil || !strings.Contains(err.Error(), "cooldown") {
		t.Fatalf("second call should fail on cooldown, got: %v", err)
	}

	// Wait for cooldown to expire
	time.Sleep(120 * time.Millisecond)

	// Third call succeeds
	if err := handler(ctx, cmdCtx); err != nil {
		t.Fatalf("third call after cooldown should succeed: %v", err)
	}

	if runs != 2 {
		t.Errorf("expected exactly 2 successful executions, got %d", runs)
	}
}

func TestMiddleware_RequirePermissions(t *testing.T) {
	adminPerms := discord.PermissionAdministrator
	userPerms := discord.PermissionSendMessages

	checkPerms := RequirePermissions(discord.PermissionManageGuild, func(cmdCtx *Context) discord.Permissions {
		if cmdCtx.Args[0] == "admin" {
			return adminPerms
		}
		return userPerms
	})

	handler := checkPerms(func(ctx context.Context, cmdCtx *Context) error {
		return nil
	})

	ctx := context.Background()

	// Admin executes
	err := handler(ctx, &Context{Args: []string{"admin"}})
	if err != nil {
		t.Errorf("admin should pass permission check: %v", err)
	}

	// Regular user executes
	err = handler(ctx, &Context{Args: []string{"user"}})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("regular user should be denied: %v", err)
	}

	_ = errors.New("")
}
