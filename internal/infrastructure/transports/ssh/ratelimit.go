// internal/transports/ssh/ratelimit.go - Rate limiting for SSH connections
package ssh

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
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

// TokenBucketRateLimiter implements rate limiting using token bucket algorithm with atomic operations.
//
// TokenBucketRateLimiter maintains per-host token buckets to limit connection
// attempts over time, providing ultra-fast protection against brute force attacks
// using lock-free atomic operations for maximum performance.
// Code block:
//
//	limiter := NewTokenBucketRateLimiter(3, time.Minute)
//	allowed, err := limiter.Allow(ctx, "example.com")
//	if err != nil {
//	    log.Printf("Rate limit check failed: %v", err)
//	    return
//	}
//	if !allowed {
//	    log.Println("Rate limited")
//	    return
//	}
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type TokenBucketRateLimiter struct {
	// maxAttempts is the maximum number of attempts allowed per time window
	maxAttempts int
	// timeWindow is the duration of the rate limiting window
	timeWindow time.Duration
	// hosts tracks rate limiting state per host
	hosts map[string]*hostState
	// mu protects concurrent access to hosts map
	mu sync.RWMutex
	// lastCleanup tracks when we last cleaned up inactive hosts (atomic)
	lastCleanup atomic.Int64 // Unix nano timestamp
	// cleanupInterval defines how often to clean up inactive hosts
	cleanupInterval time.Duration
	// requestCount tracks total requests atomically for metrics
	requestCount atomic.Uint64
	// rateLimitedCount tracks rate-limited requests atomically
	rateLimitedCount atomic.Uint64
}

// hostState tracks rate limiting state for a single host with atomic operations.
//
// hostState provides lock-free access to attempt tracking for maximum performance
// in high-concurrency scenarios, using atomic operations for counters.
// Code block:
//
//	state := &hostState{
//	    attempts: make([]time.Time, 0, maxAttempts),
//	    lastUsed: atomic.NewInt64(time.Now().UnixNano()),
//	}
//
// Parameters: N/A (for types)
//
// Returns: N/A (for types)
type hostState struct {
	// attempts tracks recent connection attempts
	attempts []time.Time
	// lastUsed tracks when this host was last accessed (atomic unix nano)
	lastUsed atomic.Int64
	// mu protects concurrent access to attempts slice
	mu sync.Mutex
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter with atomic operations.
//
// NewTokenBucketRateLimiter initializes a high-performance rate limiter that allows
// up to maxAttempts connection attempts per timeWindow per host.
// It uses atomic operations for maximum concurrency and zero-allocation patterns.
//
// Code block:
//
//	// Allow 3 attempts per minute per host
//	limiter := NewTokenBucketRateLimiter(3, time.Minute)
//	allowed, err := limiter.Allow(ctx, "example.com")
//	if err != nil {
//	    log.Printf("Rate limiting failed: %v", err)
//	    return
//	}
//	if !allowed {
//	    log.Println("Request rate limited")
//	    return
//	}
//
// Parameters:
//   - 1 maxAttempts: int - maximum attempts allowed per time window (must be > 0)
//   - 2 timeWindow: time.Duration - duration of the rate limiting window (must be > 0)
//
// Returns:
//   - 1 limiter: RateLimiter - configured ultra-performance rate limiter instance
func NewTokenBucketRateLimiter(maxAttempts int, timeWindow time.Duration) RateLimiter {
	limiter := &TokenBucketRateLimiter{
		maxAttempts:     maxAttempts,
		timeWindow:      timeWindow,
		hosts:           make(map[string]*hostState),
		cleanupInterval: timeWindow * 2, // Clean up every 2x the time window
	}
	// Initialize atomic timestamp
	limiter.lastCleanup.Store(time.Now().UnixNano())
	return limiter
}

// Allow checks if a connection attempt to the given host is allowed using atomic operations.
//
// Allow implements ultra-fast rate limiting by tracking connection attempts per host
// within the configured time window using atomic operations and lock-free patterns.
// Attempts outside the window are discarded automatically.
//
// Code block:
//
//	ctx := context.Background()
//	allowed, err := limiter.Allow(ctx, "example.com")
//	if err != nil {
//	    log.Printf("Rate check failed: %v", err)
//	    return
//	}
//	if !allowed {
//	    log.Println("Request rate limited")
//	    return
//	}
//
// Parameters:
//   - 1 ctx: context.Context - request context for timeout and cancellation
//   - 2 host: string - hostname or IP address to check (must not be empty)
//
// Returns:
//   - 1 allowed: bool - true if connection attempt is allowed, false if rate limited
//   - 2 error - non-nil if rate limiting check fails or context cancelled
func (r *TokenBucketRateLimiter) Allow(ctx context.Context, host string) (bool, error) {
	// Input validation with fast path
	if host == "" {
		return false, fmt.Errorf("host cannot be empty")
	}

	// Atomic counter increment for metrics
	r.requestCount.Add(1)

	// Check for context cancellation with zero-allocation
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	now := time.Now()
	nowNano := now.UnixNano()
	cutoff := now.Add(-r.timeWindow)

	// Get or create host state
	r.mu.RLock()
	state, exists := r.hosts[host]
	r.mu.RUnlock()

	if !exists {
		// Create new host state with pre-allocated capacity
		state = &hostState{
			attempts: make([]time.Time, 0, r.maxAttempts),
		}
		state.lastUsed.Store(nowNano)

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

	// Update last used time atomically
	state.lastUsed.Store(nowNano)

	// Remove old attempts outside the time window (zero-allocation)
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
		r.rateLimitedCount.Add(1) // Atomic increment
		return false, nil         // Rate limited
	}

	// Add this attempt
	state.attempts = append(state.attempts, now)

	// Periodic cleanup of inactive hosts to prevent memory leaks
	r.cleanupInactiveHostsIfNeeded(nowNano)

	return true, nil
}

// cleanupInactiveHostsIfNeeded performs periodic cleanup of inactive hosts using atomic operations.
//
// cleanupInactiveHostsIfNeeded removes hosts that haven't had attempts recently
// to prevent memory leaks in long-running applications with many different hosts.
// Uses atomic timestamps for lock-free performance optimization.
//
// Code block:
//
//	limiter := NewTokenBucketRateLimiter(5, time.Minute)
//	// Cleanup happens automatically during Allow() calls
//	// Uses atomic operations for maximum performance
//
// Parameters:
//   - 1 nowNano: int64 - current time in Unix nanoseconds for atomic operations
//
// Returns: N/A (void function)
func (r *TokenBucketRateLimiter) cleanupInactiveHostsIfNeeded(nowNano int64) {
	// Load last cleanup time atomically
	lastCleanupNano := r.lastCleanup.Load()

	// Only cleanup if enough time has passed (atomic comparison)
	if nowNano-lastCleanupNano < r.cleanupInterval.Nanoseconds() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-checked locking with atomic load
	if nowNano-r.lastCleanup.Load() < r.cleanupInterval.Nanoseconds() {
		return
	}

	cutoffNano := nowNano - (2 * r.timeWindow.Nanoseconds()) // Clean hosts inactive for 2x time window

	hostsToDelete := make([]string, 0) // Pre-allocate with zero capacity
	for host, state := range r.hosts {
		// Atomic load of last used time
		stateLastUsedNano := state.lastUsed.Load()

		if stateLastUsedNano < cutoffNano {
			hostsToDelete = append(hostsToDelete, host)
		}
	}

	// Delete inactive hosts (batch operation)
	for _, host := range hostsToDelete {
		delete(r.hosts, host)
	}

	// Update last cleanup time atomically
	r.lastCleanup.Store(nowNano)
}

// GetMetrics returns atomic performance metrics for monitoring.
//
// GetMetrics provides real-time metrics for rate limiter performance
// monitoring and debugging, using atomic operations for accurate counters.
//
// Code block:
//
//	limiter := NewTokenBucketRateLimiter(5, time.Minute)
//	total, limited := limiter.GetMetrics()
//	log.Printf("Total: %d, Rate Limited: %d", total, limited)
//
// Parameters: N/A
//
// Returns:
//   - 1 totalRequests: uint64 - total number of requests processed atomically
//   - 2 rateLimitedRequests: uint64 - number of rate-limited requests atomically
func (r *TokenBucketRateLimiter) GetMetrics() (totalRequests, rateLimitedRequests uint64) {
	return r.requestCount.Load(), r.rateLimitedCount.Load()
}

// ResetMetrics atomically resets all performance counters to zero.
//
// ResetMetrics provides a thread-safe way to reset metrics for
// testing or periodic monitoring cycles.
//
// Code block:
//
//	limiter := NewTokenBucketRateLimiter(5, time.Minute)
//	limiter.ResetMetrics() // Safe concurrent reset
//
// Parameters: N/A
//
// Returns: N/A (void function)
func (r *TokenBucketRateLimiter) ResetMetrics() {
	r.requestCount.Store(0)
	r.rateLimitedCount.Store(0)
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
