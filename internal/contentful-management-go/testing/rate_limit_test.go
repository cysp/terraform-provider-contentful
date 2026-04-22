package cmtesting_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	rateLimitSecondLimitHeader     = "X-Contentful-Ratelimit-Second-Limit"
	rateLimitSecondRemainingHeader = "X-Contentful-Ratelimit-Second-Remaining"
	rateLimitResetHeader           = "X-Contentful-Ratelimit-Reset"
)

func makeRateLimitTestRequest() *http.Request {
	return httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/missing", nil)
}

func TestContentfulManagementServerUsesDefaultRateLimitHeaders(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	response := httptest.NewRecorder()
	server.ServeHTTP(response, makeRateLimitTestRequest())

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "5", response.Header().Get(rateLimitSecondLimitHeader))
	assert.Equal(t, "4", response.Header().Get(rateLimitSecondRemainingHeader))
	assert.Equal(t, "0", response.Header().Get(rateLimitResetHeader))
}

func TestContentfulManagementServerRateLimitHeadersAnd429WhenEnabled(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Unix(1_000, 0))

	server, err := cmt.NewContentfulManagementServer(
		cmt.WithRateLimitPerSecond(2),
		cmt.WithRateLimitNow(clock.Now),
	)
	require.NoError(t, err)

	firstResponse := httptest.NewRecorder()
	server.ServeHTTP(firstResponse, makeRateLimitTestRequest())

	secondResponse := httptest.NewRecorder()
	server.ServeHTTP(secondResponse, makeRateLimitTestRequest())

	thirdResponse := httptest.NewRecorder()
	server.ServeHTTP(thirdResponse, makeRateLimitTestRequest())

	assert.Equal(t, http.StatusNotFound, firstResponse.Code)
	assert.Equal(t, "2", firstResponse.Header().Get(rateLimitSecondLimitHeader))
	assert.Equal(t, "1", firstResponse.Header().Get(rateLimitSecondRemainingHeader))
	assert.Equal(t, "0", firstResponse.Header().Get(rateLimitResetHeader))

	assert.Equal(t, http.StatusNotFound, secondResponse.Code)
	assert.Equal(t, "2", secondResponse.Header().Get(rateLimitSecondLimitHeader))
	assert.Equal(t, "0", secondResponse.Header().Get(rateLimitSecondRemainingHeader))
	assert.Equal(t, "0", secondResponse.Header().Get(rateLimitResetHeader))

	assert.Equal(t, http.StatusTooManyRequests, thirdResponse.Code)
	assert.Equal(t, "2", thirdResponse.Header().Get(rateLimitSecondLimitHeader))
	assert.Equal(t, "0", thirdResponse.Header().Get(rateLimitSecondRemainingHeader))
	assert.Equal(t, "1", thirdResponse.Header().Get(rateLimitResetHeader))

	var responseBody struct {
		Sys struct {
			ID string `json:"id"`
		} `json:"sys"`
	}

	err = json.Unmarshal(thirdResponse.Body.Bytes(), &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "RateLimitExceeded", responseBody.Sys.ID)
}

func TestContentfulManagementServerRateLimitResetsAfterOneSecond(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Unix(2_000, 0))

	server, err := cmt.NewContentfulManagementServer(
		cmt.WithRateLimitPerSecond(1),
		cmt.WithRateLimitNow(clock.Now),
	)
	require.NoError(t, err)

	firstResponse := httptest.NewRecorder()
	server.ServeHTTP(firstResponse, makeRateLimitTestRequest())

	secondResponse := httptest.NewRecorder()
	server.ServeHTTP(secondResponse, makeRateLimitTestRequest())

	assert.Equal(t, http.StatusNotFound, firstResponse.Code)
	assert.Equal(t, http.StatusTooManyRequests, secondResponse.Code)
	assert.Equal(t, "1", secondResponse.Header().Get(rateLimitResetHeader))

	clock.Advance(time.Second)

	afterResetResponse := httptest.NewRecorder()
	server.ServeHTTP(afterResetResponse, makeRateLimitTestRequest())

	assert.Equal(t, http.StatusNotFound, afterResetResponse.Code)
	assert.Equal(t, "0", afterResetResponse.Header().Get(rateLimitResetHeader))
}

func TestWithRateLimitPerSecondRejectsNonPositiveValues(t *testing.T) {
	t.Parallel()

	_, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(0))
	require.Error(t, err)
}

func TestWithRateLimitNowRejectsNilClock(t *testing.T) {
	t.Parallel()

	_, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitNow(nil))
	require.Error(t, err)
}
