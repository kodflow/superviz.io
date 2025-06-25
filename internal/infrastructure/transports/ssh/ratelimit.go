// internal/transports/ssh/ratelimit.go - Rate limiting for SSH connections
package ssh

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter provides rate limiting for SSH connection attempts.
//
// RateLimiter helps prevent brute force attacks by limiting the number
// of connection attempts per host within a time window.
type RateLimiter interface {
	// Allow checks if a connection attempt to the given host is allowed.
	//
	// Parameters:
	//   - ctx: context.Context for timeout and cancellation
	//   - host: string hostname or IP address to check
	//
	// Returns:
	//   - bool true if connection attempt is allowed
	//   - error if rate limiting check fails
	Allow(ctx context.Context, host string) (bool, error)
}

// TokenBucketRateLimiter implements rate limiting using token bucket algorithm.
//
// TokenBucketRateLimiter maintains per-host token buckets to limit connection
// attempts over time, providing protection against brute force attacks.
type TokenBucketRateLimiter struct {
	// maxAttempts is the maximum number of attempts allowed per time window
	maxAttempts int
	// timeWindow is the duration of the rate limiting window
	timeWindow time.Duration
	// hosts tracks rate limiting state per host
	hosts map[string]*hostState
	// mu protects concurrent access to hosts map
	mu sync.RWMutex
}

// hostState tracks rate limiting state for a single host.
type hostState struct {
	// attempts tracks recent connection attempts
	attempts []time.Time
	// mu protects concurrent access to attempts slice
	mu sync.Mutex
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter.
//
// NewTokenBucketRateLimiter initializes a rate limiter that allows
// up to maxAttempts connection attempts per timeWindow per host.
//
// Example:
//
//	// Allow 3 attempts per minute per host
//	limiter := NewTokenBucketRateLimiter(3, time.Minute)
//
// Parameters:
//   - maxAttempts: int maximum attempts allowed per time window
//   - timeWindow: time.Duration duration of the rate limiting window
//
// Returns:
//   - RateLimiter configured rate limiter instance
func NewTokenBucketRateLimiter(maxAttempts int, timeWindow time.Duration) RateLimiter {
	return &TokenBucketRateLimiter{
		maxAttempts: maxAttempts,
		timeWindow:  timeWindow,
		hosts:       make(map[string]*hostState),
	}
}

// Allow checks if a connection attempt to the given host is allowed.
//
// Allow implements rate limiting by tracking connection attempts per host
// within the configured time window. Attempts outside the window are discarded.
//
// Parameters:
//   - ctx: context.Context for timeout and cancellation
//   - host: string hostname or IP address to check
//
// Returns:
//   - bool true if connection attempt is allowed, false if rate limited
//   - error if rate limiting check fails
func (r *TokenBucketRateLimiter) Allow(ctx context.Context, host string) (bool, error) {
	if host == "" {
		return false, fmt.Errorf("host cannot be empty")
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	now := time.Now()
	cutoff := now.Add(-r.timeWindow)

	// Get or create host state
	r.mu.RLock()
	state, exists := r.hosts[host]
	r.mu.RUnlock()

	if !exists {
		// Create new host state
		state = &hostState{
			attempts: make([]time.Time, 0, r.maxAttempts),
		}

		r.mu.Lock()
		// Double-check in case another goroutine created it
		if existing, exists := r.hosts[host]; exists {
			state = existing
		} else {
			r.hosts[host] = state
		}
		r.mu.Unlock()
	}

	// Check and update attempts for this host
	state.mu.Lock()
	defer state.mu.Unlock()

	// Remove old attempts outside the time window
	validAttempts := 0
	for i, attempt := range state.attempts {
		if attempt.After(cutoff) {
			if validAttempts != i {
				state.attempts[validAttempts] = attempt
			}
			validAttempts++
		}
	}
	state.attempts = state.attempts[:validAttempts]

	// Check if we're under the limit
	if len(state.attempts) >= r.maxAttempts {
		return false, nil // Rate limited
	}

	// Add this attempt
	state.attempts = append(state.attempts, now)
	return true, nil
}

// NoOpRateLimiter is a rate limiter that always allows connections.
//
// NoOpRateLimiter can be used to disable rate limiting in development
// or testing environments.
type NoOpRateLimiter struct{}

// NewNoOpRateLimiter creates a rate limiter that never blocks.
//
// Returns:
//   - RateLimiter instance that always allows connections
func NewNoOpRateLimiter() RateLimiter {
	return &NoOpRateLimiter{}
}

// Allow always returns true, allowing all connection attempts.
//
// Parameters:
//   - ctx: context.Context (unused)
//   - host: string (unused)
//
// Returns:
//   - bool always true
//   - error always nil
func (r *NoOpRateLimiter) Allow(ctx context.Context, host string) (bool, error) {
	return true, nil
}
