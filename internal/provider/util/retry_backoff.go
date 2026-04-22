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
	rateLimitJitterMin                      = 100 * time.Millisecond
	rateLimitJitterMax                      = 300 * time.Millisecond
	maxRateLimitResetSecondsForDuration     = int64(math.MaxInt64) / int64(time.Second)
)

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

	delay := time.Duration(resetSeconds)*time.Second + contentfulRateLimitJitter()

	return applyMinBackoffFloor(delay, minDelay)
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

func contentfulRateLimitJitter() time.Duration {
	jitterRangeMilliseconds := int64((rateLimitJitterMax-rateLimitJitterMin)/time.Millisecond) + 1

	randomJitterMilliseconds, err := rand.Int(rand.Reader, big.NewInt(jitterRangeMilliseconds))
	if err != nil {
		return rateLimitJitterMin
	}

	return rateLimitJitterMin + time.Duration(randomJitterMilliseconds.Int64())*time.Millisecond
}

func applyMinBackoffFloor(delay, minDelay time.Duration) time.Duration {
	if delay < minDelay {
		return minDelay
	}

	return delay
}
