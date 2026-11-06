package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentfulRetryPolicy(t *testing.T) {
	t.Parallel()

	readMethods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	mutationMethods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, method := range readMethods {
		assertContentfulRetryPolicy(t, method+" transport error", method, 0, errContentfulRetryTestConnectionLost, true)
		assertContentfulRetryPolicy(t, method+" server error", method, http.StatusBadGateway, nil, true)
	}

	for _, method := range mutationMethods {
		assertContentfulRetryPolicy(t, method+" transport error", method, 0, errContentfulRetryTestConnectionLost, false)
		assertContentfulRetryPolicy(t, method+" server error", method, http.StatusServiceUnavailable, nil, false)
	}

	for _, method := range append(readMethods, mutationMethods...) {
		assertContentfulRetryPolicy(t, method+" rate limit", method, http.StatusTooManyRequests, nil, true)
	}

	for _, method := range mutationMethods {
		assertContentfulRetryPolicy(t, method+" rate limit with transport error", method, http.StatusTooManyRequests, errContentfulRetryTestConnectionLost, false)
	}

	assertContentfulRetryPolicy(t, "non-retryable read response", http.MethodGet, http.StatusBadRequest, nil, false)
}

func TestContentfulRetryPolicyStopsForAlreadyCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	ctx = context.WithValue(ctx, contentfulRequestMethodContextKey{}, http.MethodGet)

	shouldRetry, err := contentfulRetryPolicy(ctx, nil, errContentfulRetryTestConnectionLost)

	assert.False(t, shouldRetry)
	require.ErrorIs(t, err, context.Canceled)
}
