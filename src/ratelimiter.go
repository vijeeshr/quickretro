package main

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a per-IP token-bucket rate limiter with automatic cleanup of stale entries.
type RateLimiter struct {
	mu              sync.Mutex
	visitors        map[string]*bucket
	burst           int
	refillInterval  time.Duration
	cleanupInterval time.Duration
	stopCh          chan struct{}
}

type bucket struct {
	tokens   int
	lastSeen time.Time
}

// NewRateLimiter creates a new per-IP rate limiter.
//   - burst: maximum tokens per IP (i.e. max requests before throttling)
//   - refillInterval: how often one token is added back per IP
//   - cleanupInterval: how often stale entries are purged
func NewRateLimiter(burst int, refillInterval, cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors:        make(map[string]*bucket),
		burst:           burst,
		refillInterval:  refillInterval,
		cleanupInterval: cleanupInterval,
		stopCh:          make(chan struct{}),
	}

	go rl.cleanupLoop()

	return rl
}

// Allow checks whether a request from the given IP is allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	b, exists := rl.visitors[ip]
	if !exists {
		// First request from this IP — start with burst-1 tokens (this request consumes one)
		rl.visitors[ip] = &bucket{tokens: rl.burst - 1, lastSeen: now}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.lastSeen)
	refillCount := int(elapsed / rl.refillInterval)
	if refillCount > 0 {
		b.tokens = min(b.tokens+refillCount, rl.burst)
		b.lastSeen = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			return
		}
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.cleanupInterval)
	for ip, b := range rl.visitors {
		if b.lastSeen.Before(cutoff) {
			delete(rl.visitors, ip)
		}
	}
}

// Stop shuts down the cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// RateLimitMiddleware wraps an http.HandlerFunc with per-IP rate limiting.
// Returns 429 Too Many Requests when the limit is exceeded.
func RateLimitMiddleware(rl *RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // Fallback
		}

		if !rl.Allow(ip) {
			slog.Warn("Rate limit exceeded", "ip", ip, "path", r.URL.Path)
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

// ClientRateLimiter is a lightweight per-connection token-bucket rate limiter.
// It is NOT thread-safe — designed to be used by a single goroutine (the client's read loop).
type ClientRateLimiter struct {
	tokens         int
	burst          int
	refillInterval time.Duration
	lastRefill     time.Time
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
