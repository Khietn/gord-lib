package command

import (
	"context"
	"fmt"
	"strings"
)

// Command is the base interface for all executable commands (Command Pattern).
type Command interface {
	Name() string
	Description() string
	Execute(ctx context.Context, cmdCtx *Context) error
}

// Group implements the Composite Pattern, allowing commands to contain subcommands.
// A Group behaves as a Command itself while delegating execution to nested child commands.
type Group struct {
	name        string
	description string
	subcommands map[string]Command
}

// NewGroup constructs a new composite command group.
func NewGroup(name, description string) *Group {
	return &Group{
		name:        name,
		description: description,
		subcommands: make(map[string]Command),
	}
}

// Name returns the root name of the command group.
func (g *Group) Name() string {
	return g.name
}

// Description returns the description of the command group.
func (g *Group) Description() string {
	return g.description
}

// RegisterSubcommands attaches child subcommands to this group.
func (g *Group) RegisterSubcommands(cmds ...Command) {
	for _, cmd := range cmds {
		g.subcommands[strings.ToLower(cmd.Name())] = cmd
	}
}

// Subcommands returns the list of registered child subcommands.
func (g *Group) Subcommands() []Command {
	list := make([]Command, 0, len(g.subcommands))
	for _, cmd := range g.subcommands {
		list = append(list, cmd)
	}
	return list
}

// Execute routes execution to the appropriate child subcommand.
func (g *Group) Execute(ctx context.Context, cmdCtx *Context) error {
	if len(cmdCtx.Args) == 0 {
		return fmt.Errorf("missing subcommand for '%s'. Available: %s", g.name, g.availableSubcommands())
	}

	subName := strings.ToLower(cmdCtx.Args[0])
	subCmd, found := g.subcommands[subName]
	if !found {
		return fmt.Errorf("unknown subcommand '%s' for '%s'. Available: %s", subName, g.name, g.availableSubcommands())
	}

	// Shift arguments for child subcommand
	childCtx := &Context{
		Client:      cmdCtx.Client,
		Message:     cmdCtx.Message,
		Interaction: cmdCtx.Interaction,
		Args:        cmdCtx.Args[1:],
	}
	if len(childCtx.Args) > 0 {
		childCtx.subcommand = childCtx.Args[0]
	}

	return subCmd.Execute(ctx, childCtx)
}

func (g *Group) availableSubcommands() string {
	keys := make([]string, 0, len(g.subcommands))
	for k := range g.subcommands {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
