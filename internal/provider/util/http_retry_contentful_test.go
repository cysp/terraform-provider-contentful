package util

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContentfulRetryBackoffRateLimitedWithResetHeader(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			contentfulRateLimitResetHeader: []string{"1"},
		},
	}

	backoff := ContentfulRetryBackoff(time.Second, 30*time.Second, 3, resp)

	assert.GreaterOrEqual(t, backoff, time.Second)
	assert.LessOrEqual(t, backoff, time.Second+100*time.Millisecond)
}

func TestContentfulRetryBackoffRateLimitedWithInvalidResetHeaderFallsBack(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			contentfulRateLimitResetHeader: []string{"invalid"},
		},
	}

	backoff := ContentfulRetryBackoff(100*time.Millisecond, 200*time.Millisecond, 1, resp)

	assert.GreaterOrEqual(t, backoff, 200*time.Millisecond)
	assert.LessOrEqual(t, backoff, 400*time.Millisecond)
}

func TestContentfulRetryBackoffRateLimitedWithoutResetHeaderFallsBack(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{},
	}

	backoff := ContentfulRetryBackoff(100*time.Millisecond, 200*time.Millisecond, 1, resp)

	assert.GreaterOrEqual(t, backoff, 200*time.Millisecond)
	assert.LessOrEqual(t, backoff, 400*time.Millisecond)
}

func TestContentfulRetryBackoffNonRateLimitUsesFallback(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
	}

	backoff := ContentfulRetryBackoff(100*time.Millisecond, 200*time.Millisecond, 1, resp)

	assert.GreaterOrEqual(t, backoff, 200*time.Millisecond)
	assert.LessOrEqual(t, backoff, 400*time.Millisecond)
}
