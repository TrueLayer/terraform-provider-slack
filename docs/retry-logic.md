# Retry Logic and Rate Limiting

This document explains how the retry mechanism works and how to configure it.

## Overview

The provider automatically handles retries for all API operations. The retry logic is implemented in the `handle_retry.go` file and provides consistent behavior across all resources and data sources.

## How It Works

### 1. Rate Limiting Detection

The provider detects rate limiting through multiple mechanisms:

- **Slack RateLimitedError**: Detects Slack's native rate limiting errors
- **HTTP 429 Status**: Detects HTTP "Too Many Requests" responses
- **Retry-After Header**: Extracts the `Retry-After` header from HTTP responses

### 2. Retry Behavior

When a rate limit is detected:

1. **Logging**: The provider logs an INFO message indicating the rate limit and retry delay
2. **Wait**: The provider waits for the specified retry time (from `Retry-After` header or Slack's rate limit error)
3. **Retry**: The operation is automatically retried after the wait period
4. **Completion**: When the retry completes, another INFO message is logged

### 3. Transient Error Handling

The provider also retries on transient errors such as:
- Network timeouts
- Connection refused errors
- Server errors (5xx status codes)
- Temporary failures

## Configuration

### Provider Configuration

You can configure the retry timeout in your provider block:

```hcl
provider "slack" {
  token = var.slack_token
  retry_timeout = 300  # 5 minutes (default: 60 seconds)
}
```

### Environment Variable

You can also set the retry timeout via environment variable:

```bash
export TF_VAR_retry_timeout=300
```

## Logging

The provider logs retry events at the INFO level:

```
2025-07-07T10:02:57.188+0100 [INFO]  provider.terraform-provider-slack_v1.0.0: Slack rate limit exceeded, retrying after delay: @module=provider error="slack rate limit exceeded, retry after 10s" retry_after_seconds=10 tf_req_id=41122934-6F8D-450B-9514-AA54B5B30C69 tf_resource_type=slack_conversation [...] tf_rpc=ReadResource timestamp="2025-07-07T10:02:57.188+0100"
2025-07-07T10:03:07.190+0100 [INFO]  provider.terraform-provider-slack_v1.0.0: Rate limit wait completed, retrying operation: [...] tf_req_id=41122934-6F8D-450B-9514-AA54B5B30C69 tf_resource_type=slack_conversation tf_rpc=ReadResource @module=provider tf_provider_addr=provider timestamp="2025-07-07T10:03:07.190+0100"
```

## Error Handling

### Retryable vs Non-Retryable Errors

- **Retryable**: Rate limits, network errors, server errors
- **Non-Retryable**: Authentication errors, validation errors, context cancellation

### Context Cancellation

If the context is cancelled during a retry wait, the operation is immediately aborted and returns a non-retryable error.
