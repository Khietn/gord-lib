package rest

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Bucket represents a single rate-limiting bucket tracked for a specific route.
type Bucket struct {
	mu        sync.Mutex
	limit     int
	remaining int
	resetAt   time.Time
}

// Wait blocks until a request token is available in this bucket or context is cancelled.
func (b *Bucket) Wait(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	// If reset time has passed, reset remaining to at least 1
	if now.After(b.resetAt) {
		b.remaining = 1
	}

	if b.remaining <= 0 {
		sleepDuration := b.resetAt.Sub(now)
		if sleepDuration > 0 {
			timer := time.NewTimer(sleepDuration)
			defer timer.Stop()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				b.remaining = 1
			}
		}
	}

	b.remaining--
	return nil
}

// Update refreshes bucket statistics using Discord response headers.
func (b *Bucket) Update(remaining int, resetAt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.remaining = remaining
	b.resetAt = resetAt
}

// RateLimiter manages route-based rate limits and global rate limits for Discord API requests.
type RateLimiter struct {
	mu          sync.RWMutex
	globalUntil time.Time
	buckets     map[string]*Bucket
	routeHashes map[string]string
}

// NewRateLimiter creates a new RateLimiter instance.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets:     make(map[string]*Bucket),
		routeHashes: make(map[string]string),
	}
}

// getBucketIdentifier calculates the bucket key given a route.
func (rl *RateLimiter) getBucketIdentifier(route *Route) string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	routeKey := route.BucketKey + ":" + route.MajorParam
	if hash, ok := rl.routeHashes[routeKey]; ok {
		return hash + ":" + route.MajorParam
	}
	return routeKey
}

// getOrCreateBucket returns an existing bucket or creates a fresh one.
func (rl *RateLimiter) getOrCreateBucket(key string) *Bucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if b, ok := rl.buckets[key]; ok {
		return b
	}

	b := &Bucket{
		limit:     1,
		remaining: 1,
		resetAt:   time.Now(),
	}
	rl.buckets[key] = b
	return b
}

// Wait pauses execution if a global rate limit is active or if the route's bucket is exhausted.
func (rl *RateLimiter) Wait(ctx context.Context, route *Route) error {
	// Check global rate limit
	rl.mu.RLock()
	globalSleep := time.Until(rl.globalUntil)
	rl.mu.RUnlock()

	if globalSleep > 0 {
		timer := time.NewTimer(globalSleep)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}

	bucketKey := rl.getBucketIdentifier(route)
	bucket := rl.getOrCreateBucket(bucketKey)
	return bucket.Wait(ctx)
}

// Update updates rate limiter state based on Discord HTTP response headers.
func (rl *RateLimiter) Update(route *Route, headers http.Header) {
	bucketHash := headers.Get("X-RateLimit-Bucket")
	remainingStr := headers.Get("X-RateLimit-Remaining")
	resetAfterStr := headers.Get("X-RateLimit-Reset-After")
	isGlobal := headers.Get("X-RateLimit-Global") == "true"

	if isGlobal {
		retryAfterStr := headers.Get("Retry-After")
		if seconds, err := strconv.ParseFloat(retryAfterStr, 64); err == nil {
			rl.mu.Lock()
			rl.globalUntil = time.Now().Add(time.Duration(seconds * float64(time.Second)))
			rl.mu.Unlock()
		}
		return
	}

	if resetAfterStr != "" && remainingStr != "" {
		remaining, errRem := strconv.Atoi(remainingStr)
		resetAfter, errReset := strconv.ParseFloat(resetAfterStr, 64)

		if errRem == nil && errReset == nil {
			resetAt := time.Now().Add(time.Duration(resetAfter * float64(time.Second)))

			if bucketHash != "" {
				routeKey := route.BucketKey + ":" + route.MajorParam
				rl.mu.Lock()
				rl.routeHashes[routeKey] = bucketHash
				rl.mu.Unlock()

				hashKey := bucketHash + ":" + route.MajorParam
				bucket := rl.getOrCreateBucket(hashKey)
				bucket.Update(remaining, resetAt)
			} else {
				routeKey := route.BucketKey + ":" + route.MajorParam
				bucket := rl.getOrCreateBucket(routeKey)
				bucket.Update(remaining, resetAt)
			}
		}
	}
}
