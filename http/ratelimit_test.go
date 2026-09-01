package fbhttp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoginRateLimiter(t *testing.T) {
	limiter := NewLoginRateLimiter(3, time.Minute, 100*time.Millisecond)
	key := "127.0.0.1"

	assert.True(t, limiter.Allow(key))

	limiter.RecordFailure(key)
	assert.True(t, limiter.Allow(key))

	limiter.RecordFailure(key)
	assert.True(t, limiter.Allow(key))

	limiter.RecordFailure(key) // 3rd failure -> locked out
	assert.False(t, limiter.Allow(key))

	// Wait for lockout duration
	time.Sleep(120 * time.Millisecond)
	assert.True(t, limiter.Allow(key))

	// Reset test
	limiter.RecordFailure(key)
	limiter.Reset(key)
	limiter.RecordFailure(key)
	assert.True(t, limiter.Allow(key))
}
