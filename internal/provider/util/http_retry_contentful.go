package util

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const contentfulRateLimitResetHeader = "X-Contentful-RateLimit-Reset"

const (
	maxRateLimitHeaderJitter = 250 * time.Millisecond
)

// ContentfulRetryBackoff applies Contentful rate-limit reset delays when available.
// For 429 responses without a usable reset header it falls back to retryablehttp.LinearJitterBackoff.
func ContentfulRetryBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if resetSeconds, ok := contentfulRateLimitResetSeconds(resp.Header); ok {
			wait := time.Duration(resetSeconds) * time.Second

			return wait + positiveJitter(wait)
		}
	}

	return retryablehttp.LinearJitterBackoff(min, max, attemptNum, resp)
}

func contentfulRateLimitResetSeconds(header http.Header) (int64, bool) {
	if header == nil {
		return 0, false
	}

	resetRaw := ""
	for headerName, headerValues := range header {
		if strings.EqualFold(headerName, contentfulRateLimitResetHeader) && len(headerValues) > 0 {
			resetRaw = strings.TrimSpace(headerValues[0])

			break
		}
	}

	if resetRaw == "" {
		return 0, false
	}

	resetSeconds, err := strconv.ParseInt(resetRaw, 10, 64)
	if err != nil || resetSeconds <= 0 {
		return 0, false
	}

	return resetSeconds, true
}

func positiveJitter(wait time.Duration) time.Duration {
	if wait <= 0 {
		return 0
	}

	jitter := min(wait/10, maxRateLimitHeaderJitter)

	if jitter <= 0 {
		return 0
	}

	return time.Duration(rand.Int63n(int64(jitter) + 1))
}
