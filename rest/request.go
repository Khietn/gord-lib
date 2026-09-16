package rest

import (
	"fmt"
)

// DiscordError represents an error response returned by the Discord REST API.
type DiscordError struct {
	StatusCode int    `json:"status_code"`
	Code       int    `json:"code"`
	Message    string `json:"message"`
}

// Error formats the Discord error into a human-readable string.
func (e *DiscordError) Error() string {
	return fmt.Sprintf("discord rest error: status=%d, code=%d, message=%s", e.StatusCode, e.Code, e.Message)
}

// DefaultUserAgent is the library user-agent sent to Discord.
const DefaultUserAgent = "DiscordBot (https://github.com/Khietn/gord-lib, 0.1.0)"

// APIRoot is the Discord REST API v10 base endpoint.
const APIRoot = "https://discord.com/api/v10"
