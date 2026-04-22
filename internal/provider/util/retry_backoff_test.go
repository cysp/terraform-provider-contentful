package util_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/stretchr/testify/assert"
)

func TestContentfulRateLimitLinearJitterBackoffUsesContentfulResetHeader(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"X-Contentful-RateLimit-Reset": []string{"1"},
		},
	}

	delay := util.ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

	assert.GreaterOrEqual(t, delay, 1100*time.Millisecond)
	assert.LessOrEqual(t, delay, 1300*time.Millisecond)
}

func TestContentfulRateLimitLinearJitterBackoffUsesLargeContentfulResetHeader(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"X-Contentful-RateLimit-Reset": []string{"900"},
		},
	}

	delay := util.ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

	assert.GreaterOrEqual(t, delay, 900*time.Second+100*time.Millisecond)
	assert.LessOrEqual(t, delay, 900*time.Second+300*time.Millisecond)
}

func TestContentfulRateLimitLinearJitterBackoffAppliesMinDelayFloorForZeroResetHeader(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"X-Contentful-RateLimit-Reset": []string{"0"},
		},
	}

	delay := util.ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

	assert.Equal(t, time.Second, delay)
}

func TestContentfulRateLimitLinearJitterBackoffFallsBackToRetryAfterOnInvalidContentfulResetHeader(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"X-Contentful-RateLimit-Reset": []string{"invalid"},
			"Retry-After":                  []string{"2"},
		},
	}

	delay := util.ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

	assert.Equal(t, 2*time.Second, delay)
}

func TestContentfulRateLimitLinearJitterBackoffFallsBackToRetryAfterOnOversizedContentfulResetHeader(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"X-Contentful-RateLimit-Reset": []string{"9223372036854775807"},
			"Retry-After":                  []string{"2"},
		},
	}

	delay := util.ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

	assert.Equal(t, 2*time.Second, delay)
}

func TestContentfulRateLimitLinearJitterBackoffAcceptsRatelimitHeaderCasingVariant(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header: http.Header{
			"X-Contentful-Ratelimit-Reset": []string{"1"},
		},
	}

	delay := util.ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

	assert.GreaterOrEqual(t, delay, 1100*time.Millisecond)
	assert.LessOrEqual(t, delay, 1300*time.Millisecond)
}

func TestContentfulRateLimitLinearJitterBackoffDelegatesToFallbackForNon429Responses(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Header: http.Header{
			"Retry-After": []string{"3"},
		},
	}

	delay := util.ContentfulRateLimitLinearJitterBackoff(time.Second, 3*time.Second, 0, resp)

	assert.Equal(t, 3*time.Second, delay)
}
