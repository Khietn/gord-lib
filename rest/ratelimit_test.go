package rest

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Khietn/gord-lib/discord"
)

func TestRateLimiter_BucketWait(t *testing.T) {
	rl := NewRateLimiter()
	route := RouteCreateMessage(discord.Snowflake(123456))

	// Initial wait should be instantaneous
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := rl.Wait(ctx, route); err != nil {
		t.Fatalf("first wait should succeed: %v", err)
	}

	// Update with remaining = 0 and reset_after = 0.2s
	headers := make(http.Header)
	headers.Set("X-RateLimit-Bucket", "test-bucket")
	headers.Set("X-RateLimit-Remaining", "0")
	headers.Set("X-RateLimit-Reset-After", "0.2")

	rl.Update(route, headers)

	// Second wait with very short timeout should fail because bucket is exhausted
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer shortCancel()

	err := rl.Wait(shortCtx, route)
	if err == nil {
		t.Errorf("expected wait to block/timeout on exhausted bucket, but got nil")
	}

	// Third wait with enough time should succeed once reset passes
	longCtx, longCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer longCancel()

	if err := rl.Wait(longCtx, route); err != nil {
		t.Errorf("expected wait to succeed after reset, got: %v", err)
	}
}
