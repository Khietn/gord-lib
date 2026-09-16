package command

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Khietn/gord-lib"
	"github.com/Khietn/gord-lib/events"
)

// Router acts as the Invoker and Central Registry in the Command Pattern.
type Router struct {
	mu          sync.RWMutex
	prefix      string
	commands    map[string]Command
	middlewares []Middleware
}

// NewRouter constructs a new command Router.
// If no prefix is specified, it defaults to "/" (standard slash command prefix).
func NewRouter(prefix ...string) *Router {
	p := "/"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return &Router{
		prefix:   p,
		commands: make(map[string]Command),
	}
}

// Prefix returns the configured command prefix (default "/").
func (r *Router) Prefix() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.prefix
}

// Use appends one or more global middlewares to the execution chain.
func (r *Router) Use(middlewares ...Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middlewares...)
}

// Register registers one or more commands with the router.
func (r *Router) Register(cmds ...Command) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, cmd := range cmds {
		r.commands[strings.ToLower(cmd.Name())] = cmd
	}
}

// Find retrieves a registered command by name.
func (r *Router) Find(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, ok := r.commands[strings.ToLower(name)]
	return cmd, ok
}

// Commands returns a slice of all registered commands.
func (r *Router) Commands() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		list = append(list, cmd)
	}
	return list
}

// HandleMessage parses and routes an incoming Discord MessageCreate event.
// Validates that the message begins with the designated prefix (e.g. "/").
func (r *Router) HandleMessage(ctx context.Context, client *gord.Client, e events.MessageCreate) {
	if e.Author.Bot || !strings.HasPrefix(e.Content, r.prefix) {
		return
	}

	content := strings.TrimPrefix(e.Content, r.prefix)
	fields := strings.Fields(content)
	if len(fields) == 0 {
		return
	}

	cmdName := strings.ToLower(fields[0])
	args := fields[1:]

	cmd, found := r.Find(cmdName)
	if !found {
		return
	}

	cmdCtx := NewMessageContext(client, &e, args)
	if err := r.execute(ctx, cmd, cmdCtx); err != nil {
		_, _ = cmdCtx.Reply(ctx, fmt.Sprintf("⚠️ %v", err))
	}
}

// execute runs the command through the registered middleware pipeline.
func (r *Router) execute(ctx context.Context, cmd Command, cmdCtx *Context) error {
	r.mu.RLock()
	middlewares := make([]Middleware, len(r.middlewares))
	copy(middlewares, r.middlewares)
	r.mu.RUnlock()

	handler := func(c context.Context, cc *Context) error {
		return cmd.Execute(c, cc)
	}

	// Build the middleware chain (Chain of Responsibility)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler(ctx, cmdCtx)
}
