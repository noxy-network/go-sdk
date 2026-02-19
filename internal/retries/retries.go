package retries

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Retryable gRPC status codes: UNKNOWN, DEADLINE_EXCEEDED, RESOURCE_EXHAUSTED, UNAVAILABLE
var retryableCodes = map[codes.Code]bool{
	codes.Unknown:            true,
	codes.DeadlineExceeded:    true,
	codes.ResourceExhausted:  true,
	codes.Unavailable:        true,
}

// IsRetryable returns true if the error indicates a retryable network/unavailable condition.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return retryableCodes[s.Code()]
}

// WithRetry executes fn up to retries times, with exponential backoff on retryable errors.
func WithRetry[T any](ctx context.Context, fn func() (T, error), retries int) (T, error) {
	var zero T
	var lastErr error
	for i := 0; i < retries; i++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
		if i < retries-1 && IsRetryable(err) {
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(time.Duration(1<<i) * 100 * time.Millisecond):
				// continue
			}
		} else {
			break
		}
	}
	return zero, lastErr
}
