package cmtesting

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	xContentfulRateLimitSecondLimitHeader     = "X-Contentful-Ratelimit-Second-Limit"
	xContentfulRateLimitSecondRemainingHeader = "X-Contentful-Ratelimit-Second-Remaining"
	xContentfulRateLimitResetHeader           = "X-Contentful-Ratelimit-Reset"
)

type rateLimitState struct {
	allowed   bool
	limit     int
	remaining int
	reset     int
}

func (s rateLimitState) writeHeaders(header http.Header) {
	header.Set(xContentfulRateLimitSecondLimitHeader, strconv.Itoa(s.limit))
	header.Set(xContentfulRateLimitSecondRemainingHeader, strconv.Itoa(s.remaining))
	header.Set(xContentfulRateLimitResetHeader, strconv.Itoa(s.reset))
}

type secondRateLimiter struct {
	limit int
	now   func() time.Time

	mu           sync.Mutex
	windowSecond int64
	usedInWindow int
}

func newSecondRateLimiter(limit int, now func() time.Time) *secondRateLimiter {
	return &secondRateLimiter{
		limit: limit,
		now:   now,
	}
}

func (l *secondRateLimiter) currentRateLimitState() rateLimitState {
	if l == nil {
		return rateLimitState{
			allowed:   true,
			limit:     0,
			remaining: 0,
			reset:     0,
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	currentSecond := l.now().Unix()
	if currentSecond != l.windowSecond {
		l.windowSecond = currentSecond
		l.usedInWindow = 0
	}

	if l.usedInWindow >= l.limit {
		return rateLimitState{
			allowed:   false,
			limit:     l.limit,
			remaining: 0,
			reset:     1,
		}
	}

	l.usedInWindow++

	return rateLimitState{
		allowed:   true,
		limit:     l.limit,
		remaining: l.limit - l.usedInWindow,
		reset:     0,
	}
}

func (s *Server) currentRateLimitState() rateLimitState {
	return s.rateLimiter.currentRateLimitState()
}
