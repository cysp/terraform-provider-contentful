package util

import (
	"crypto/rand"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	contentfulRateLimitResetHeader          = "X-Contentful-RateLimit-Reset"
	contentfulRatelimitResetCanonicalHeader = "X-Contentful-Ratelimit-Reset"
	rateLimitPostResetFloor                 = 100 * time.Millisecond
	rateLimitContentionWindowBase           = 500 * time.Millisecond
	rateLimitContentionWindowMultiplier     = 2
	rateLimitContentionWindowMax            = 4 * time.Second
	maxRateLimitResetSecondsForDuration     = int64(math.MaxInt64-rateLimitPostResetFloor-rateLimitContentionWindowMax) / int64(time.Second)
)

// ContentfulRateLimitLinearJitterBackoff treats Contentful's valid reset value
// as the earliest retry time, then adds a 100ms floor and bounded exponential
// full jitter for contention. Missing or invalid reset values retain
// retryablehttp's Retry-After and linear-jitter fallback behavior.
func ContentfulRateLimitLinearJitterBackoff(minDelay, maxDelay time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp == nil || resp.StatusCode != http.StatusTooManyRequests {
		return retryablehttp.RateLimitLinearJitterBackoff(minDelay, maxDelay, attemptNum, resp)
	}

	resetSeconds, ok := parseContentfulRateLimitReset(resp.Header)
	if !ok {
		return retryablehttp.RateLimitLinearJitterBackoff(minDelay, maxDelay, attemptNum, resp)
	}

	if resetSeconds > maxRateLimitResetSecondsForDuration {
		return retryablehttp.RateLimitLinearJitterBackoff(minDelay, maxDelay, attemptNum, resp)
	}

	contentionWindow := contentfulRateLimitContentionWindow(attemptNum)
	jitter := contentfulRateLimitFullJitter(contentionWindow)

	return contentfulRateLimitDelay(resetSeconds, jitter)
}

func parseContentfulRateLimitReset(headers http.Header) (int64, bool) {
	if headers == nil {
		return 0, false
	}

	reset := firstHeaderValue(headers, contentfulRatelimitResetCanonicalHeader, contentfulRateLimitResetHeader)
	if reset == "" {
		return 0, false
	}

	resetSeconds, err := strconv.ParseInt(reset, 10, 64)
	if err != nil || resetSeconds < 0 {
		return 0, false
	}

	return resetSeconds, true
}

func firstHeaderValue(headers http.Header, keys ...string) string {
	for _, key := range keys {
		if value := headers.Get(key); value != "" {
			return value
		}

		if values, ok := headers[key]; ok && len(values) > 0 && values[0] != "" {
			return values[0]
		}
	}

	return ""
}

func contentfulRateLimitContentionWindow(attemptNum int) time.Duration {
	window := rateLimitContentionWindowBase

	for attemptNum > 0 && window < rateLimitContentionWindowMax {
		window = min(window*rateLimitContentionWindowMultiplier, rateLimitContentionWindowMax)
		attemptNum--
	}

	return window
}

func contentfulRateLimitDelay(resetSeconds int64, jitter time.Duration) time.Duration {
	return time.Duration(resetSeconds)*time.Second + rateLimitPostResetFloor + jitter
}

func contentfulRateLimitFullJitter(window time.Duration) time.Duration {
	randomJitter, err := rand.Int(rand.Reader, big.NewInt(int64(window)))
	if err != nil {
		return 0
	}

	return time.Duration(randomJitter.Int64())
}
