package slack

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	assert.NotNil(t, config)
	assert.Equal(t, DefaultRetryTimeout(), config.Timeout)
}

func TestWithRetry_Success(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{Timeout: 1 * time.Second}

	called := false
	err := WithRetry(ctx, config, func() error {
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{Timeout: 1 * time.Second}

	expectedErr := errors.New("permanent error")
	called := false
	err := WithRetry(ctx, config, func() error {
		called = true
		return expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.True(t, called)
}

func TestWithRetry_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &RetryConfig{Timeout: 5 * time.Second}

	cancel() // Cancel immediately

	err := WithRetry(ctx, config, func() error {
		return errors.New("some error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestWithRetryWithResult_Success(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{Timeout: 1 * time.Second}

	expectedResult := "test result"
	called := false
	result, err := WithRetryWithResult(ctx, config, func() (string, error) {
		called = true
		return expectedResult, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.True(t, called)
}

func TestWithRetryWithResult_Error(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{Timeout: 1 * time.Second}

	expectedErr := errors.New("test error")
	called := false
	result, err := WithRetryWithResult(ctx, config, func() (string, error) {
		called = true
		return "", expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Empty(t, result)
	assert.True(t, called)
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "timeout error",
			err:      errors.New("timeout"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "network error",
			err:      errors.New("network error"),
			expected: true,
		},
		{
			name:     "server error",
			err:      errors.New("server error"),
			expected: true,
		},
		{
			name:     "internal server error",
			err:      errors.New("internal server error"),
			expected: true,
		},
		{
			name:     "service unavailable",
			err:      errors.New("service unavailable"),
			expected: true,
		},
		{
			name:     "bad gateway",
			err:      errors.New("bad gateway"),
			expected: true,
		},
		{
			name:     "gateway timeout",
			err:      errors.New("gateway timeout"),
			expected: true,
		},
		{
			name:     "permanent error",
			err:      errors.New("permanent error"),
			expected: false,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: false,
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsSubstring(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "contains substring",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "does not contain substring",
			s:        "hello world",
			substr:   "mars",
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "test",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "test",
			substr:   "",
			expected: true,
		},
		{
			name:     "exact match",
			s:        "test",
			substr:   "test",
			expected: true,
		},
		{
			name:     "case sensitive",
			s:        "Hello World",
			substr:   "world",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsSubstring(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithRetry_RateLimitError(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{Timeout: 2 * time.Second}

	rateLimitErr := &slack.RateLimitedError{
		RetryAfter: 100 * time.Millisecond,
	}

	callCount := 0
	err := WithRetry(ctx, config, func() error {
		callCount++
		if callCount == 1 {
			return rateLimitErr
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestWithRetryWithResult_RateLimitError(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{Timeout: 2 * time.Second}

	rateLimitErr := &slack.RateLimitedError{
		RetryAfter: 100 * time.Millisecond,
	}

	callCount := 0
	result, err := WithRetryWithResult(ctx, config, func() (string, error) {
		callCount++
		if callCount == 1 {
			return "", rateLimitErr
		}
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 2, callCount)
}
