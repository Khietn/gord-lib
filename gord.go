package gord

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Khietn/gord-lib/discord"
	"github.com/Khietn/gord-lib/events"
	"github.com/Khietn/gord-lib/gateway"
	"github.com/Khietn/gord-lib/rest"
)

var _ io.Closer = (*Client)(nil)

// Client is the high-level facade tying REST, Gateway, and Events together.
type Client struct {
	token         string
	intents       discord.GatewayIntents
	applicationID discord.Snowflake
	publicKey     string

	Rest       *rest.Client
	Gateway    *gateway.Gateway
	Dispatcher *events.Dispatcher
	dialer     gateway.Dialer
}

// Option configures the Client facade.
type Option func(*Client)

// WithIntents sets the Discord Gateway intents.
func WithIntents(intents discord.GatewayIntents) Option {
	return func(c *Client) {
		c.intents = intents
	}
}

// WithApplicationID sets the bot's application ID.
func WithApplicationID(appID discord.Snowflake) Option {
	return func(c *Client) {
		c.applicationID = appID
	}
}

// WithPublicKey sets the Ed25519 public key for interaction verification.
func WithPublicKey(pubKey string) Option {
	return func(c *Client) {
		c.publicKey = pubKey
	}
}

// WithGatewayDialer provides a custom WebSocket dialer implementation.
func WithGatewayDialer(dialer gateway.Dialer) Option {
	return func(c *Client) {
		c.dialer = dialer
	}
}

// WithWorkerCount sets the worker pool size for handling events.
func WithWorkerCount(workers int) Option {
	return func(c *Client) {
		if c.Dispatcher != nil {
			// Replace with new dispatcher if option order dictates
			_ = c.Dispatcher.Close()
		}
		c.Dispatcher = events.NewDispatcher(events.WithWorkerCount(workers))
	}
}

// New creates and configures a new unified gord Client.
func New(token string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	dispatcher := events.NewDispatcher(events.WithWorkerCount(8))
	restClient := rest.NewClient(token)

	c := &Client{
		token:      token,
		intents:    discord.IntentsDefault,
		Rest:       restClient,
		Dispatcher: dispatcher,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.dialer == nil {
		c.dialer = gateway.NewDefaultDialer()
	}

	c.Gateway = gateway.New(gateway.Config{
		Token:      c.token,
		Intents:    c.intents,
		Dialer:     c.dialer,
		Dispatcher: c.Dispatcher,
	})

	return c, nil
}

// Open retrieves the recommended Gateway endpoint from REST and connects the WebSocket.
func (c *Client) Open(ctx context.Context) error {
	botGateway, err := c.Rest.GetGatewayBot(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gateway endpoint: %w", err)
	}

	endpoint := botGateway.URL + "/?v=10&encoding=json"
	return c.Gateway.Open(ctx, endpoint)
}

// Close gracefully terminates the Gateway connection and flushes the event worker pool.
func (c *Client) Close() error {
	gwErr := c.Gateway.Close()
	dispErr := c.Dispatcher.Close()

	if gwErr != nil {
		return gwErr
	}
	return dispErr
}

// On registers a strongly-typed event handler on the internal dispatcher.
func On[E events.Event](c *Client, handler func(ctx context.Context, e E)) {
	events.Register(c.Dispatcher, handler)
}

// HTTPInteractionHandler creates an http.HandlerFunc that verifies Ed25519 signatures
// and routes incoming interaction webhook callbacks.
func (c *Client) HTTPInteractionHandler(handler discord.InteractionHandler) (http.HandlerFunc, error) {
	if c.publicKey == "" {
		return nil, fmt.Errorf("cannot create HTTP interaction handler: public key is not configured")
	}
	return discord.NewHTTPInteractionHandler(c.publicKey, handler), nil
}

// LoadEnv parses and loads environment variables from a .env file.
// If no filenames are provided, it defaults to searching for ".env".
func LoadEnv(filenames ...string) error {
	filename := ".env"
	if len(filenames) > 0 {
		filename = filenames[0]
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
				(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
				val = val[1 : len(val)-1]
			}
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
	return scanner.Err()
}
