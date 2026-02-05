package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	MaxRequests int
	Window      time.Duration
}

// checkRateLimit checks if a user has exceeded rate limits
// Returns error if limit exceeded, nil if allowed
func checkRateLimit(ctx context.Context, key string, config RateLimitConfig) error {
	if db == nil {
		// If no DB, allow (for tests)
		return nil
	}

	// Create rate limit tracking table if not exists
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS rate_limits (
			key TEXT PRIMARY KEY,
			count INTEGER DEFAULT 0,
			window_start TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Printf("Failed to create rate_limits table: %v", err)
		return nil // Don't block on rate limit errors
	}

	// Check and increment counter
	var count int
	var windowStart time.Time

	err = db.QueryRow(ctx, `
		INSERT INTO rate_limits (key, count, window_start)
		VALUES ($1, 1, NOW())
		ON CONFLICT (key) DO UPDATE
		SET count = CASE
			WHEN rate_limits.window_start < NOW() - $2::interval THEN 1
			ELSE rate_limits.count + 1
		END,
		window_start = CASE
			WHEN rate_limits.window_start < NOW() - $2::interval THEN NOW()
			ELSE rate_limits.window_start
		END
		RETURNING count, window_start
	`, key, fmt.Sprintf("%d seconds", int(config.Window.Seconds()))).Scan(&count, &windowStart)

	if err != nil {
		log.Printf("Rate limit check error: %v", err)
		return nil // Don't block on errors
	}

	if count > config.MaxRequests {
		remaining := config.Window - time.Since(windowStart)
		return fmt.Errorf("Rate limit exceeded. Try again in %v", remaining.Round(time.Second))
	}

	return nil
}

// cleanupRateLimits removes old rate limit entries (should be called periodically)
func cleanupRateLimits(ctx context.Context) error {
	if db == nil {
		return nil
	}

	_, err := db.Exec(ctx, `
		DELETE FROM rate_limits 
		WHERE window_start < NOW() - INTERVAL '1 hour'
	`)

	return err
}
