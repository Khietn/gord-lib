package discord

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// DiscordEpoch is the Discord epoch in milliseconds (2015-01-01T00:00:00Z).
const DiscordEpoch = 1420070400000

// Snowflake represents a unique Discord entity ID (Twitter snowflake format).
type Snowflake uint64

// ParseSnowflake parses a string representation of a Discord snowflake.
func ParseSnowflake(s string) (Snowflake, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid snowflake %q: %w", s, err)
	}
	return Snowflake(v), nil
}

// String returns the string representation of the Snowflake.
func (s Snowflake) String() string {
	return strconv.FormatUint(uint64(s), 10)
}

// IsValid checks whether the snowflake is non-zero.
func (s Snowflake) IsValid() bool {
	return s != 0
}

// CreatedAt extracts the creation timestamp from the Snowflake.
func (s Snowflake) CreatedAt() time.Time {
	msec := int64((s >> 22) + DiscordEpoch)
	return time.UnixMilli(msec).UTC()
}

// WorkerID returns the internal worker ID that generated this Snowflake.
func (s Snowflake) WorkerID() byte {
	return byte((s & 0x3E0000) >> 17)
}

// ProcessID returns the internal process ID that generated this Snowflake.
func (s Snowflake) ProcessID() byte {
	return byte((s & 0x1F000) >> 12)
}

// Sequence returns the internal sequence number of this Snowflake.
func (s Snowflake) Sequence() uint16 {
	return uint16(s & 0xFFF)
}

// MarshalJSON marshals the Snowflake as a JSON string to preserve 64-bit precision.
func (s Snowflake) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(s.String())), nil
}

// UnmarshalJSON unmarshals either a string or integer JSON value into a Snowflake.
func (s *Snowflake) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*s = 0
		return nil
	}

	// If enclosed in quotes, strip them
	if data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}

	v, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to unmarshal snowflake %q: %w", string(data), err)
	}

	*s = Snowflake(v)
	return nil
}
