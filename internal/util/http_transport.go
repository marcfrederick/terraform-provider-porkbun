package util

import (
	"fmt"
	"net/http"
	"time"
)

var _ http.RoundTripper = (*retryTransport)(nil)

// retryTransport is a custom http.RoundTripper that retries requests on rate limit errors.
//
// The Porkbun API returns a 503 Service Unavailable status code when the rate
// limit is exceeded. This transport retries the request up to maxRetries
// times, using exponential backoff (1s, 2s, 4s, etc.) between attempts.
//
// If all retries are exhausted, the last error or a retry exhaustion error is returned.
type retryTransport struct {
	transport  http.RoundTripper
	maxRetries int
}

// NewRetryTransport creates a new http.RoundTripper with retry capability.
func NewRetryTransport(transport http.RoundTripper, maxRetries int) http.RoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}

	return &retryTransport{
		transport:  transport,
		maxRetries: maxRetries,
	}
}

func (r *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		resp, err := r.transport.RoundTrip(req)
		if err == nil && resp.StatusCode != http.StatusServiceUnavailable {
			return resp, nil
		}

		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("received status code %d", resp.StatusCode)
			_ = resp.Body.Close() // prevent resource leaks
		}

		if attempt == r.maxRetries {
			break
		}

		backoff := time.Second << attempt
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(backoff):
			// retry after backoff
		}
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", r.maxRetries, lastErr)
}
