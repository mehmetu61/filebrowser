package fbhttp

import (
	"net/http"
	"sync"
	"time"

	"github.com/tomasen/realip"
)

// LoginRateLimiter limits failed login attempts to prevent brute force attacks.
type LoginRateLimiter struct {
	mu           sync.Mutex
	attempts     map[string]int
	lockouts     map[string]time.Time
	maxAttempts  int
	window       time.Duration
	lockDuration time.Duration
}

// Global default login rate limiter
var defaultLoginLimiter = NewLoginRateLimiter(5, time.Minute, 5*time.Minute)

// NewLoginRateLimiter creates a new rate limiter instance.
func NewLoginRateLimiter(maxAttempts int, window, lockDuration time.Duration) *LoginRateLimiter {
	limiter := &LoginRateLimiter{
		attempts:     make(map[string]int),
		lockouts:     make(map[string]time.Time),
		maxAttempts:  maxAttempts,
		window:       window,
		lockDuration: lockDuration,
	}

	// Periodic cleanup of expired lockouts
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return limiter
}

func (l *LoginRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for k, lockUntil := range l.lockouts {
		if now.After(lockUntil) {
			delete(l.lockouts, k)
			delete(l.attempts, k)
		}
	}
}

// Allow checks if requests from this key are allowed.
func (l *LoginRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if lockUntil, ok := l.lockouts[key]; ok {
		if time.Now().Before(lockUntil) {
			return false
		}
		// Lock expired
		delete(l.lockouts, key)
		delete(l.attempts, key)
	}

	return true
}

// RecordFailure records a failed login attempt for the key.
func (l *LoginRateLimiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.attempts[key]++
	if l.attempts[key] >= l.maxAttempts {
		l.lockouts[key] = time.Now().Add(l.lockDuration)
	}
}

// Reset clears failure tracking after a successful login.
func (l *LoginRateLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.attempts, key)
	delete(l.lockouts, key)
}

// GetClientIdentifier extracts the client IP address from the request.
func GetClientIdentifier(r *http.Request) string {
	ip := realip.FromRequest(r)
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}
