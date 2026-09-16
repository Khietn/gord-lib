package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Khietn/gord-lib/discord"
)

// GatewayBotResponse represents the response from GET /gateway/bot.
type GatewayBotResponse struct {
	URL    string `json:"url"`
	Shards int    `json:"shards"`
}

// Option configures a REST Client.
type Option func(*Client)

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL overrides the default Discord API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithRateLimiter provides a custom RateLimiter instance.
func WithRateLimiter(rl *RateLimiter) Option {
	return func(c *Client) {
		c.rateLimiter = rl
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// Client represents the Discord REST API client.
type Client struct {
	token       string
	httpClient  *http.Client
	rateLimiter *RateLimiter
	baseURL     string
	userAgent   string
}

// NewClient creates a new Discord REST client with functional options.
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		token:       token,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		rateLimiter: NewRateLimiter(),
		baseURL:     APIRoot,
		userAgent:   DefaultUserAgent,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Do executes an HTTP request against the Discord REST API while honoring rate limits.
func (c *Client) Do(ctx context.Context, route *Route, body any, out any) error {
	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	reqURL := c.baseURL + route.Path
	req, err := http.NewRequestWithContext(ctx, route.Method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+c.token)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Wait on RateLimiter
	if err := c.rateLimiter.Wait(ctx, route); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Update bucket headers
	c.rateLimiter.Update(route, resp.Header)

	// Check response status
	if resp.StatusCode >= 400 {
		var discErr DiscordError
		discErr.StatusCode = resp.StatusCode
		_ = json.NewDecoder(resp.Body).Decode(&discErr)
		return &discErr
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}

	return nil
}

// CreateMessage sends a message to a Discord channel.
func (c *Client) CreateMessage(ctx context.Context, channelID discord.Snowflake, payload discord.MessageCreatePayload) (*discord.Message, error) {
	route := RouteCreateMessage(channelID)
	var msg discord.Message
	if err := c.Do(ctx, route, payload, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetChannel retrieves channel information.
func (c *Client) GetChannel(ctx context.Context, channelID discord.Snowflake) (*discord.Channel, error) {
	route := RouteGetChannel(channelID)
	var ch discord.Channel
	if err := c.Do(ctx, route, nil, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// GetGuild retrieves guild information.
func (c *Client) GetGuild(ctx context.Context, guildID discord.Snowflake) (*discord.Guild, error) {
	route := RouteGetGuild(guildID)
	var guild discord.Guild
	if err := c.Do(ctx, route, nil, &guild); err != nil {
		return nil, err
	}
	return &guild, nil
}

// CreateInteractionResponse sends a response callback to an interaction.
func (c *Client) CreateInteractionResponse(ctx context.Context, interactionID discord.Snowflake, token string, resp discord.InteractionResponse) error {
	route := RouteCreateInteractionResponse(interactionID, token)
	return c.Do(ctx, route, resp, nil)
}

// GetGatewayBot retrieves Gateway bot connection URL and recommended shards.
func (c *Client) GetGatewayBot(ctx context.Context) (*GatewayBotResponse, error) {
	route := RouteGetGatewayBot()
	var g GatewayBotResponse
	if err := c.Do(ctx, route, nil, &g); err != nil {
		return nil, err
	}
	return &g, nil
}
