package gateway

import (
	"encoding/json"

	"github.com/Khietn/gord-lib/discord"
)

// Gateway OpCodes as defined in the Discord Gateway documentation.
const (
	OpcodeDispatch            = 0
	OpcodeHeartbeat           = 1
	OpcodeIdentify            = 2
	OpcodePresenceUpdate      = 3
	OpcodeVoiceStateUpdate    = 4
	OpcodeResume              = 6
	OpcodeReconnect           = 7
	OpcodeRequestGuildMembers = 8
	OpcodeInvalidSession      = 9
	OpcodeHello               = 10
	OpcodeHeartbeatACK        = 11
)

// Payload represents a raw Discord Gateway WebSocket frame.
type Payload struct {
	Op        int             `json:"op"`
	Data      json.RawMessage `json:"d,omitempty"`
	Sequence  *int64          `json:"s,omitempty"`
	EventName string          `json:"t,omitempty"`
}

// HelloData represents the payload received in OpcodeHello (10).
type HelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// IdentifyData represents the payload sent in OpcodeIdentify (2).
type IdentifyData struct {
	Token          string                 `json:"token"`
	Properties     IdentifyProperties     `json:"properties"`
	Intents        discord.GatewayIntents `json:"intents"`
	Shard          *[2]int                `json:"shard,omitempty"`
	LargeThreshold int                    `json:"large_threshold,omitempty"`
}

// IdentifyProperties contains system information sent during Identify.
type IdentifyProperties struct {
	OS      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

// ResumeData represents the payload sent in OpcodeResume (6).
type ResumeData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Sequence  int64  `json:"seq"`
}

// ReadyData represents the payload received in the READY dispatch event.
type ReadyData struct {
	V           int          `json:"v"`
	User        discord.User `json:"user"`
	SessionID   string       `json:"session_id"`
	ResumeURL   string       `json:"resume_gateway_url"`
	Shard       *[2]int      `json:"shard,omitempty"`
	Application struct {
		ID discord.Snowflake `json:"id"`
	} `json:"application"`
}
