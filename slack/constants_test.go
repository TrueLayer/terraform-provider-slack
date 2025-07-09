package slack

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRetryTimeout(t *testing.T) {
	timeout := DefaultRetryTimeout()
	expected := time.Duration(DefaultRetryTimeoutSeconds) * time.Second

	assert.Equal(t, expected, timeout)
	assert.Equal(t, 60*time.Second, timeout)
}

func TestDefaultRetryTimeoutSeconds(t *testing.T) {
	assert.Equal(t, 60, DefaultRetryTimeoutSeconds)
}
