package ai

import (
	"context"
	"fmt"
	"time"
)

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int) *RateLimiter {
	return &RateLimiter{
		limit:     limit,
		lastReset: time.Now(),
		tokens:    0,
	}
}

// CheckLimit checks if the current request would exceed rate limits
func (r *RateLimiter) CheckLimit() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.lastReset) >= time.Minute {
		r.tokens = 0
		r.lastReset = now
	}

	if r.tokens >= r.limit {
		return fmt.Errorf("request rate limit exceeded")
	}

	r.tokens++
	return nil
}

// AddTokens adds to the token count and checks the limit
func (r *RateLimiter) AddTokens(tokens int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tokens+tokens > r.limit {
		return fmt.Errorf("token rate limit exceeded")
	}

	r.tokens += tokens
	return nil
}

// Wait waits for the rate limiter to allow a request
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Sub(r.lastReset) >= time.Minute {
		r.tokens = 0
		r.lastReset = now
	}

	if r.tokens >= r.limit {
		return fmt.Errorf("rate limit exceeded")
	}

	r.tokens++
	return nil
}
