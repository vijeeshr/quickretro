package main

import (
	"time"
)

// ClientRateLimiter is a lightweight per-connection token-bucket rate limiter.
// It is NOT thread-safe - designed to be used by a single goroutine (the client's read loop).
type ClientRateLimiter struct {
	lastRefill     time.Time
	refillInterval time.Duration
	tokens         int
	burst          int
}

// NewClientRateLimiter creates a rate limiter for a single WebSocket client.
func NewClientRateLimiter(burst int, refillInterval time.Duration) *ClientRateLimiter {
	return &ClientRateLimiter{
		tokens:         burst,
		burst:          burst,
		refillInterval: refillInterval,
		lastRefill:     time.Now(),
	}
}

// Allow checks if the client is within the rate limit. NOT thread-safe.
func (cl *ClientRateLimiter) Allow() bool {
	now := time.Now()

	// Refill tokens based on elapsed time
	elapsed := now.Sub(cl.lastRefill)
	refillCount := int(elapsed / cl.refillInterval)
	if refillCount > 0 {
		cl.tokens = min(cl.tokens+refillCount, cl.burst)
		cl.lastRefill = now
	}

	if cl.tokens > 0 {
		cl.tokens--
		return true
	}

	return false
}
