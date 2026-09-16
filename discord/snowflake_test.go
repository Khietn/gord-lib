package discord

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSnowflake_JSON(t *testing.T) {
	const rawID = "175928847299117063"
	var sf Snowflake

	// Test unmarshaling string
	err := json.Unmarshal([]byte(`"`+rawID+`"`), &sf)
	if err != nil {
		t.Fatalf("failed to unmarshal string snowflake: %v", err)
	}
	if sf.String() != rawID {
		t.Fatalf("expected snowflake string %s, got %s", rawID, sf.String())
	}

	// Test unmarshaling raw integer
	var sfInt Snowflake
	err = json.Unmarshal([]byte(rawID), &sfInt)
	if err != nil {
		t.Fatalf("failed to unmarshal int snowflake: %v", err)
	}
	if sfInt != sf {
		t.Fatalf("expected snowflake %v, got %v", sf, sfInt)
	}

	// Test marshaling
	marshaled, err := json.Marshal(sf)
	if err != nil {
		t.Fatalf("failed to marshal snowflake: %v", err)
	}
	if string(marshaled) != `"`+rawID+`"` {
		t.Fatalf("expected %q, got %q", `"`+rawID+`"`, string(marshaled))
	}
}

func TestSnowflake_Timestamp(t *testing.T) {
	// Discord official example snowflake: 175928847299117063
	sf, err := ParseSnowflake("175928847299117063")
	if err != nil {
		t.Fatalf("failed to parse snowflake: %v", err)
	}

	created := sf.CreatedAt()
	expected := time.Date(2016, 4, 30, 11, 18, 25, 796000000, time.UTC)
	if !created.Equal(expected) {
		t.Fatalf("expected timestamp %v, got %v", expected, created)
	}

	if sf.WorkerID() != 1 {
		t.Fatalf("expected worker ID 1, got %d", sf.WorkerID())
	}
	if sf.ProcessID() != 0 {
		t.Fatalf("expected process ID 0, got %d", sf.ProcessID())
	}
	if sf.Sequence() != 7 {
		t.Fatalf("expected sequence 7, got %d", sf.Sequence())
	}
}
