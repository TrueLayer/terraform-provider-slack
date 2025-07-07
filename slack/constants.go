package slack

import "time"

const (
	// DefaultRetryTimeoutSeconds is the default retry timeout in seconds
	DefaultRetryTimeoutSeconds = 60
)

// DefaultRetryTimeout returns the default retry timeout as a duration
func DefaultRetryTimeout() time.Duration {
	return time.Duration(DefaultRetryTimeoutSeconds) * time.Second
}
