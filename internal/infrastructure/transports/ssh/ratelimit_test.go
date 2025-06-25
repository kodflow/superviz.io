// internal/transports/ssh/ratelimit_test.go - Tests for SSH rate limiting
package ssh

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenBucketRateLimiter(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(3, time.Minute)
	require.NotNil(t, limiter)

	// Check type assertion
	bucketLimiter, ok := limiter.(*TokenBucketRateLimiter)
	require.True(t, ok)
	assert.Equal(t, 3, bucketLimiter.maxAttempts)
	assert.Equal(t, time.Minute, bucketLimiter.timeWindow)
	assert.NotNil(t, bucketLimiter.hosts)
}

func TestTokenBucketRateLimiter_Allow_Success(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(3, time.Minute)
	ctx := context.Background()
	host := "example.com"

	// First 3 attempts should be allowed
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, host)
		require.NoError(t, err)
		assert.True(t, allowed, "attempt %d should be allowed", i+1)
	}

	// 4th attempt should be rate limited
	allowed, err := limiter.Allow(ctx, host)
	require.NoError(t, err)
	assert.False(t, allowed, "4th attempt should be rate limited")
}

func TestTokenBucketRateLimiter_Allow_DifferentHosts(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(2, time.Minute)
	ctx := context.Background()

	// Each host should have its own rate limit
	allowed1, err := limiter.Allow(ctx, "host1.example.com")
	require.NoError(t, err)
	assert.True(t, allowed1)

	allowed2, err := limiter.Allow(ctx, "host2.example.com")
	require.NoError(t, err)
	assert.True(t, allowed2)

	// Fill up host1 quota
	allowed1, err = limiter.Allow(ctx, "host1.example.com")
	require.NoError(t, err)
	assert.True(t, allowed1)

	// host1 should be rate limited
	allowed1, err = limiter.Allow(ctx, "host1.example.com")
	require.NoError(t, err)
	assert.False(t, allowed1)

	// host2 should still be allowed
	allowed2, err = limiter.Allow(ctx, "host2.example.com")
	require.NoError(t, err)
	assert.True(t, allowed2)
}

func TestTokenBucketRateLimiter_Allow_TimeWindow(t *testing.T) {
	// Use very short time window for testing
	limiter := NewTokenBucketRateLimiter(1, 50*time.Millisecond)
	ctx := context.Background()
	host := "example.com"

	// First attempt should be allowed
	allowed, err := limiter.Allow(ctx, host)
	require.NoError(t, err)
	assert.True(t, allowed)

	// Second attempt should be rate limited
	allowed, err = limiter.Allow(ctx, host)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Wait for time window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again after time window
	allowed, err = limiter.Allow(ctx, host)
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestTokenBucketRateLimiter_Allow_EmptyHost(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(3, time.Minute)
	ctx := context.Background()

	allowed, err := limiter.Allow(ctx, "")
	require.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "host cannot be empty")
}

func TestTokenBucketRateLimiter_Allow_ContextCancellation(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(3, time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	allowed, err := limiter.Allow(ctx, "example.com")
	require.Error(t, err)
	assert.False(t, allowed)
	assert.Equal(t, context.Canceled, err)
}

func TestTokenBucketRateLimiter_Allow_ContextTimeout(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(3, time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(1 * time.Millisecond)

	allowed, err := limiter.Allow(ctx, "example.com")
	require.Error(t, err)
	assert.False(t, allowed)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestTokenBucketRateLimiter_Allow_ConcurrentAccess(t *testing.T) {
	limiter := NewTokenBucketRateLimiter(10, time.Minute)
	ctx := context.Background()
	host := "example.com"

	// Test concurrent access doesn't cause race conditions
	done := make(chan bool, 20)
	var allowedCount int32

	for i := 0; i < 20; i++ {
		go func() {
			defer func() { done <- true }()
			allowed, err := limiter.Allow(ctx, host)
			require.NoError(t, err)
			if allowed {
				atomic.AddInt32(&allowedCount, 1)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should have allowed exactly 10 attempts (our limit)
	assert.Equal(t, int32(10), atomic.LoadInt32(&allowedCount))
}

func TestNoOpRateLimiter(t *testing.T) {
	limiter := NewNoOpRateLimiter()
	require.NotNil(t, limiter)

	ctx := context.Background()

	// Should always allow, regardless of how many attempts
	for i := 0; i < 100; i++ {
		allowed, err := limiter.Allow(ctx, "example.com")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// Should work with empty host too
	allowed, err := limiter.Allow(ctx, "")
	require.NoError(t, err)
	assert.True(t, allowed)

	// Should work with cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	allowed, err = limiter.Allow(cancelledCtx, "example.com")
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestTokenBucketRateLimiter_Allow_MemoryCleanup(t *testing.T) {
	// Test that the cleanup function exists and doesn't panic
	limiter := NewTokenBucketRateLimiter(2, 50*time.Millisecond)
	ctx := context.Background()
	host := "example.com"

	// Fill up the bucket
	allowed, err := limiter.Allow(ctx, host)
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = limiter.Allow(ctx, host)
	require.NoError(t, err)
	assert.True(t, allowed)

	// Verify that the cleanup function can be called without panic
	rateLimiter := limiter.(*TokenBucketRateLimiter)
	rateLimiter.cleanupInactiveHostsIfNeeded(time.Now())

	// Host should still exist since it was just used
	rateLimiter.mu.RLock()
	_, exists := rateLimiter.hosts[host]
	rateLimiter.mu.RUnlock()
	assert.True(t, exists, "Host should still exist since it was just used")
}

func BenchmarkTokenBucketRateLimiter_Allow(b *testing.B) {
	limiter := NewTokenBucketRateLimiter(1000000, time.Hour) // High limit to avoid rate limiting
	ctx := context.Background()
	host := "example.com"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = limiter.Allow(ctx, host)
		}
	})
}

func BenchmarkNoOpRateLimiter_Allow(b *testing.B) {
	limiter := NewNoOpRateLimiter()
	ctx := context.Background()
	host := "example.com"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = limiter.Allow(ctx, host)
		}
	})
}
