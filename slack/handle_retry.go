package slack

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/slack-go/slack"
)

// RetryConfig holds the retry configuration for the provider
type RetryConfig struct {
	Timeout time.Duration
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		Timeout: 60 * time.Second,
	}
}

// WithRetry executes a function with retry logic for rate limiting and other transient errors
func WithRetry(ctx context.Context, config *RetryConfig, operation func() error) error {
	return retry.RetryContext(ctx, config.Timeout, func() *retry.RetryError {
		err := operation()
		if err == nil {
			return nil
		}

		// Handle Slack rate limiting
		var rateLimitErr *slack.RateLimitedError
		if errors.As(err, &rateLimitErr) {
			retryAfter := rateLimitErr.RetryAfter
			tflog.Info(ctx, "Slack rate limit exceeded, retrying after delay", map[string]interface{}{
				"retry_after_seconds": retryAfter.Seconds(),
				"error":               err.Error(),
			})

			// Wait for the specified retry time
			select {
			case <-ctx.Done():
				return retry.NonRetryableError(fmt.Errorf("context canceled during rate limit wait: %w", ctx.Err()))
			case <-time.After(retryAfter):
				tflog.Info(ctx, "Rate limit wait completed, retrying operation")
				return retry.RetryableError(err)
			}
		}

		// Note: HTTP 429 handling is not implemented as the slack-go library
		// doesn't expose the HTTP response directly. The RateLimitedError
		// should handle most rate limiting scenarios.

		// Handle other transient errors that might be retryable
		if isRetryableError(err) {
			tflog.Info(ctx, "Transient error detected, retrying operation", map[string]interface{}{
				"error": err.Error(),
			})
			return retry.RetryableError(err)
		}

		// Non-retryable error
		return retry.NonRetryableError(err)
	})
}

// WithRetryWithResult executes a function with retry logic and returns a result
func WithRetryWithResult[T any](ctx context.Context, config *RetryConfig, operation func() (T, error)) (T, error) {
	var result T
	err := retry.RetryContext(ctx, config.Timeout, func() *retry.RetryError {
		var opErr error
		result, opErr = operation()
		if opErr == nil {
			return nil
		}

		// Handle Slack rate limiting
		var rateLimitErr *slack.RateLimitedError
		if errors.As(opErr, &rateLimitErr) {
			retryAfter := rateLimitErr.RetryAfter
			tflog.Info(ctx, "Slack rate limit exceeded, retrying after delay", map[string]interface{}{
				"retry_after_seconds": retryAfter.Seconds(),
				"error":               opErr.Error(),
			})

			select {
			case <-ctx.Done():
				return retry.NonRetryableError(fmt.Errorf("context canceled during rate limit wait: %w", ctx.Err()))
			case <-time.After(retryAfter):
				tflog.Info(ctx, "Rate limit wait completed, retrying operation")
				return retry.RetryableError(opErr)
			}
		}

		// Note: HTTP 429 handling is not implemented as the slack-go library
		// doesn't expose the HTTP response directly. The RateLimitedError
		// should handle most rate limiting scenarios.

		// Handle other transient errors that might be retryable
		if isRetryableError(opErr) {
			tflog.Info(ctx, "Transient error detected, retrying operation", map[string]interface{}{
				"error": opErr.Error(),
			})
			return retry.RetryableError(opErr)
		}

		// Non-retryable error
		return retry.NonRetryableError(opErr)
	})

	return result, err
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common transient errors
	errStr := err.Error()

	// Network-related errors
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false // These are context errors, not retryable
	}

	// Slack-specific transient errors
	transientErrors := []string{
		"timeout",
		"connection refused",
		"network error",
		"temporary failure",
		"server error",
		"internal server error",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
	}

	for _, transientErr := range transientErrors {
		if containsSubstring(errStr, transientErr) {
			return true
		}
	}

	return false
}

// containsSubstring performs a simple substring search
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
