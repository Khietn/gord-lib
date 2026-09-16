package gateway

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestHeartbeatManager_BeatAndAck(t *testing.T) {
	var beatsCount int
	var mu sync.Mutex
	var hm *HeartbeatManager

	sendBeat := func(seq *int64) error {
		mu.Lock()
		beatsCount++
		mu.Unlock()
		if hm != nil {
			hm.Acknowledge()
		}
		return nil
	}

	zombieCalled := false
	onZombie := func() {
		zombieCalled = true
	}

	hm = NewHeartbeatManager(40, sendBeat, onZombie)
	defer hm.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	seqVal := int64(42)
	hm.Start(ctx, func() *int64 { return &seqVal })

	// Allow first beat and ACK to happen
	time.Sleep(60 * time.Millisecond)

	lat := hm.Latency()
	if lat < 0 {
		t.Errorf("latency must not be negative, got %v", lat)
	}

	mu.Lock()
	count := beatsCount
	mu.Unlock()

	if count == 0 {
		t.Errorf("expected at least 1 beat sent, got %d", count)
	}

	if zombieCalled {
		t.Errorf("zombie callback should not be triggered when acknowledged")
	}
}

func TestHeartbeatManager_ZombieDetection(t *testing.T) {
	zombieCh := make(chan struct{}, 1)

	// Never ACK to trigger zombie callback
	sendBeat := func(seq *int64) error {
		return nil
	}

	hm := NewHeartbeatManager(20, sendBeat, func() {
		select {
		case zombieCh <- struct{}{}:
		default:
		}
	})
	defer hm.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hm.Start(ctx, func() *int64 { return nil })

	select {
	case <-zombieCh:
		// Zombie detected successfully
	case <-time.After(500 * time.Millisecond):
		t.Errorf("timed out waiting for zombie connection detection")
	}
}

func TestMakeHeartbeatPayload(t *testing.T) {
	seq := int64(100)
	raw, err := MakeHeartbeatPayload(&seq)
	if err != nil {
		t.Fatalf("MakeHeartbeatPayload failed: %v", err)
	}

	var p Payload
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if p.Op != OpcodeHeartbeat {
		t.Errorf("expected OpcodeHeartbeat (1), got %d", p.Op)
	}

	// Test nil sequence
	nilRaw, err := MakeHeartbeatPayload(nil)
	if err != nil {
		t.Fatalf("MakeHeartbeatPayload(nil) failed: %v", err)
	}
	var nilP Payload
	if err := json.Unmarshal(nilRaw, &nilP); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if nilP.Op != OpcodeHeartbeat {
		t.Errorf("expected OpcodeHeartbeat, got %d", nilP.Op)
	}
}
