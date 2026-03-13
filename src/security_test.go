package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --------------------
// isOriginAllowed tests
// --------------------

func TestIsOriginAllowed_AllowedOrigin(t *testing.T) {
	// Setup: set config with allowed origins
	origConfig := config
	t.Cleanup(func() { config = origConfig })

	config.Server.AllowedOrigins = []string{"http://localhost:8080", "https://example.com"}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "http://localhost:8080")

	if !isOriginAllowed(req) {
		t.Error("expected origin 'http://localhost:8080' to be allowed")
	}
}

func TestIsOriginAllowed_DisallowedOrigin(t *testing.T) {
	origConfig := config
	t.Cleanup(func() { config = origConfig })

	config.Server.AllowedOrigins = []string{"http://localhost:8080"}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "https://evil-site.com")

	if isOriginAllowed(req) {
		t.Error("expected origin 'https://evil-site.com' to be rejected")
	}
}

func TestIsOriginAllowed_EmptyOrigin(t *testing.T) {
	origConfig := config
	t.Cleanup(func() { config = origConfig })

	config.Server.AllowedOrigins = []string{"http://localhost:8080"}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	// No Origin header set

	if isOriginAllowed(req) {
		t.Error("expected empty origin to be rejected")
	}
}

// --------------------
// Security headers middleware tests
// --------------------

func TestSecurityHeaders_SetsAllHeaders(t *testing.T) {
	env := EnvironmentConfig{TurnstileEnabled: false}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := securityHeaders(env, inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	expected := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":          "DENY",
		"Referrer-Policy":          "strict-origin-when-cross-origin",
		"Permissions-Policy":       "camera=(), microphone=(), geolocation=()",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
	}

	for header, want := range expected {
		got := rr.Header().Get(header)
		if got != want {
			t.Errorf("header %q = %q, want %q", header, got, want)
		}
	}

	// CSP should be present
	csp := rr.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("expected Content-Security-Policy header to be set")
	}
}

func TestSecurityHeaders_CSPIncludesTurnstile_WhenEnabled(t *testing.T) {
	env := EnvironmentConfig{TurnstileEnabled: true}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := securityHeaders(env, inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	csp := rr.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("expected Content-Security-Policy to be set")
	}

	// Check that Turnstile domain is included in CSP
	if !containsSubstring(csp, "https://challenges.cloudflare.com") {
		t.Errorf("CSP should include challenges.cloudflare.com when Turnstile is enabled, got: %s", csp)
	}
}

// --------------------
// RateLimiter tests
// --------------------

func TestRateLimiter_AllowsWithinBurst(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute, time.Hour)
	defer rl.Stop()

	for i := 0; i < 3; i++ {
		if !rl.Allow("127.0.0.1") {
			t.Errorf("request %d should be allowed within burst of 3", i+1)
		}
	}
}

func TestRateLimiter_BlocksAfterBurst(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute, time.Hour)
	defer rl.Stop()

	rl.Allow("127.0.0.1") // 1st — ok
	rl.Allow("127.0.0.1") // 2nd — ok

	if rl.Allow("127.0.0.1") {
		t.Error("3rd request should be blocked after burst of 2")
	}
}

func TestRateLimiter_RefillsAfterInterval(t *testing.T) {
	// Use a very short refill interval for testing
	rl := NewRateLimiter(1, 10*time.Millisecond, time.Hour)
	defer rl.Stop()

	rl.Allow("127.0.0.1") // Uses the single token

	if rl.Allow("127.0.0.1") {
		t.Error("should be blocked immediately after exhausting burst")
	}

	// Wait for refill
	time.Sleep(15 * time.Millisecond)

	if !rl.Allow("127.0.0.1") {
		t.Error("should be allowed after refill interval")
	}
}

func TestRateLimiter_TracksIPsSeparately(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute, time.Hour)
	defer rl.Stop()

	rl.Allow("10.0.0.1") // Exhausts IP1's token

	if !rl.Allow("10.0.0.2") {
		t.Error("different IP should have its own token bucket")
	}
}

func TestRateLimitMiddleware_Returns429(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute, time.Hour)
	defer rl.Stop()

	handler := RateLimitMiddleware(rl, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First request — allowed
	req1 := httptest.NewRequest(http.MethodPost, "/", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("first request: expected status 200, got %d", rr1.Code)
	}

	// Second request — rate limited
	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected status 429, got %d", rr2.Code)
	}
}

// --------------------
// Helper
// --------------------

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringImpl(s, substr))
}

func containsSubstringImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --------------------
// ClientRateLimiter tests
// --------------------

func TestClientRateLimiter_AllowsWithinBurst(t *testing.T) {
	cl := NewClientRateLimiter(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !cl.Allow() {
			t.Errorf("request %d should be allowed within burst of 5", i+1)
		}
	}
}

func TestClientRateLimiter_BlocksAfterBurst(t *testing.T) {
	cl := NewClientRateLimiter(2, time.Minute)
	cl.Allow()
	cl.Allow()
	if cl.Allow() {
		t.Error("3rd message should be blocked after burst of 2")
	}
}

func TestClientRateLimiter_RefillsAfterInterval(t *testing.T) {
	cl := NewClientRateLimiter(1, 10*time.Millisecond)
	cl.Allow() // Exhaust
	if cl.Allow() {
		t.Error("should be blocked")
	}
	time.Sleep(15 * time.Millisecond)
	if !cl.Allow() {
		t.Error("should be allowed after refill")
	}
}
