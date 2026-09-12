package util //nolint:testpackage // deterministic tests exercise unexported pure delay arithmetic.

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContentfulRateLimitContentionWindowWidensAndCaps(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		attemptNum int
		window     time.Duration
	}{
		"initial attempt": {attemptNum: 0, window: 500 * time.Millisecond},
		"second attempt":  {attemptNum: 1, window: time.Second},
		"third attempt":   {attemptNum: 2, window: 2 * time.Second},
		"cap reached":     {attemptNum: 3, window: 4 * time.Second},
		"cap retained":    {attemptNum: 20, window: 4 * time.Second},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.window, contentfulRateLimitContentionWindow(test.attemptNum))
		})
	}
}

func TestContentfulRateLimitDelayUsesResetAsLowerBoundWithoutMultiplication(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 904*time.Second+99*time.Millisecond, contentfulRateLimitDelay(900, 3999*time.Millisecond))
}

func TestContentfulRateLimitLinearJitterBackoffDoesNotMultiplyLargeReset(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"X-Contentful-RateLimit-Reset": []string{"900"},
		},
	}

	delay := ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 20, resp)

	assert.GreaterOrEqual(t, delay, 900*time.Second+100*time.Millisecond)
	assert.Less(t, delay, 904*time.Second+100*time.Millisecond)
}

func TestContentfulRateLimitDelayPreservesPositivePostResetFloor(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 100*time.Millisecond, contentfulRateLimitDelay(0, 0))
}

func TestContentfulRateLimitLinearJitterBackoffFallsBackForUnusableContentfulReset(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"missing":   "",
		"invalid":   "invalid",
		"negative":  "-1",
		"oversized": "9223372036854775807",
	}

	for name, reset := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			header := http.Header{"Retry-After": []string{"2"}}
			if reset != "" {
				header.Set("X-Contentful-Ratelimit-Reset", reset)
			}

			resp := &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     header,
			}

			delay := ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

			assert.Equal(t, 2*time.Second, delay)
		})
	}
}

func TestContentfulRateLimitLinearJitterBackoffAcceptsRatelimitHeaderCasingVariant(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"X-Contentful-Ratelimit-Reset": []string{"1"},
		},
	}

	delay := ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

	assert.GreaterOrEqual(t, delay, 1100*time.Millisecond)
	assert.Less(t, delay, 1600*time.Millisecond)
}

func TestContentfulRateLimitLinearJitterBackoffDelegatesToFallbackForNon429Responses(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Header: http.Header{
			"Retry-After": []string{"3"},
		},
	}

	delay := ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

	assert.Equal(t, 3*time.Second, delay)
}
