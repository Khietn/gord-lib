package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

// HeartbeatManager maintains the heartbeat loop with Discord Gateway.
type HeartbeatManager struct {
	interval  time.Duration
	lastSent  atomic.Int64
	lastAcked atomic.Int64
	acked     atomic.Bool
	stopCh    chan struct{}
	sendBeat  func(seq *int64) error
	onZombie  func()
}

// NewHeartbeatManager creates a new heartbeat coordinator.
func NewHeartbeatManager(intervalMs int, sendBeat func(seq *int64) error, onZombie func()) *HeartbeatManager {
	hm := &HeartbeatManager{
		interval: time.Duration(intervalMs) * time.Millisecond,
		stopCh:   make(chan struct{}),
		sendBeat: sendBeat,
		onZombie: onZombie,
	}
	hm.acked.Store(true) // Start assuming ready for first ACK
	return hm
}

// Start initiates the heartbeat loop honoring jitter.
func (hm *HeartbeatManager) Start(ctx context.Context, getSeq func() *int64) {
	// First heartbeat includes jitter: interval * rand(0, 1)
	jitterRatio := rand.Float64()
	initialDelay := time.Duration(float64(hm.interval) * jitterRatio)

	go func() {
		timer := time.NewTimer(initialDelay)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return
		case <-hm.stopCh:
			return
		case <-timer.C:
			// Send first heartbeat
			hm.beat(getSeq())
		}

		ticker := time.NewTicker(hm.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-hm.stopCh:
				return
			case <-ticker.C:
				// Zombie connection check: if last beat wasn't acked, connection is dead
				if !hm.acked.Load() {
					if hm.onZombie != nil {
						hm.onZombie()
					}
					return
				}
				hm.beat(getSeq())
			}
		}
	}()
}

// beat sends an OpcodeHeartbeat payload with the latest sequence.
func (hm *HeartbeatManager) beat(seq *int64) {
	hm.acked.Store(false)
	hm.lastSent.Store(time.Now().UnixMilli())
	if err := hm.sendBeat(seq); err != nil {
		fmt.Printf("[gord-lib/gateway] error sending heartbeat: %v\n", err)
	}
}

// Acknowledge records receipt of OpcodeHeartbeatACK from Discord.
func (hm *HeartbeatManager) Acknowledge() {
	hm.acked.Store(true)
	hm.lastAcked.Store(time.Now().UnixMilli())
}

// Stop terminates the heartbeat goroutine.
func (hm *HeartbeatManager) Stop() {
	select {
	case <-hm.stopCh:
	default:
		close(hm.stopCh)
	}
}

// Latency calculates the round-trip latency of the last heartbeat in milliseconds.
func (hm *HeartbeatManager) Latency() time.Duration {
	sent := hm.lastSent.Load()
	acked := hm.lastAcked.Load()
	if acked >= sent && sent > 0 {
		return time.Duration(acked-sent) * time.Millisecond
	}
	return 0
}

// MakeHeartbeatPayload formats an OpcodeHeartbeat payload.
func MakeHeartbeatPayload(seq *int64) ([]byte, error) {
	p := Payload{
		Op: OpcodeHeartbeat,
	}
	if seq != nil {
		raw, _ := json.Marshal(*seq)
		p.Data = raw
	}
	return json.Marshal(p)
}
