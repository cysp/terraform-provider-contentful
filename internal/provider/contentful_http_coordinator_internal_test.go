package provider

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentfulRetryCoordinatorLogsResponseRetryDecisions(t *testing.T) {
	t.Parallel()

	const (
		targetSentinel = "https://token.test.invalid/private?access_token=URL_SENTINEL"
		headerSentinel = "HEADER_SENTINEL"
		bodySentinel   = "BODY_SENTINEL"
		tokenSentinel  = "TOKEN_SENTINEL"
	)

	tests := map[string]struct {
		deadline     time.Duration
		status       int
		requestErr   error
		shouldRetry  bool
		decision     string
		expectsEvent bool
	}{
		"accepted response retry":             {deadline: 5 * time.Second, status: http.StatusTooManyRequests, shouldRetry: true, decision: "retry", expectsEvent: true},
		"response retry declined at deadline": {deadline: time.Second, status: http.StatusTooManyRequests, decision: "decline_deadline", expectsEvent: true},
		"transport error":                     {requestErr: errContentfulRetryTestLogSentinel, shouldRetry: true},
		"non-retryable response":              {status: http.StatusBadRequest},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var logOutput bytes.Buffer

			ctx := tflogtest.RootLogger(t.Context(), &logOutput)

			if test.deadline > 0 {
				var cancel context.CancelFunc

				ctx, cancel = context.WithTimeout(ctx, test.deadline)
				t.Cleanup(cancel)
			}

			ctx = context.WithValue(ctx, contentfulRequestMethodContextKey{}, http.MethodGet)
			ctx = context.WithValue(ctx, contentfulRequestRetryStateContextKey{}, &contentfulRequestRetryState{})

			var response *http.Response

			if test.status != 0 {
				request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetSentinel, strings.NewReader(bodySentinel))
				require.NoError(t, err)
				request.Header.Set("Authorization", "Bearer "+tokenSentinel)
				response = contentfulRetryTestResponse(request, test.status)
				response.Header.Set("Retry-After", "2")
				response.Header.Set("X-Test-Secret", headerSentinel)
				response.Body = io.NopCloser(strings.NewReader(bodySentinel))

				t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
			}

			retryCoordinator := contentfulRetryCoordinator{client: retryablehttp.NewClient()}
			shouldRetry, err := retryCoordinator.checkRetry(ctx, response, test.requestErr)
			require.NoError(t, err)
			assert.Equal(t, test.shouldRetry, shouldRetry)

			entries, err := tflogtest.MultilineJSONDecode(bytes.NewReader(logOutput.Bytes()))
			require.NoError(t, err)

			if test.expectsEvent {
				require.Len(t, entries, 1)
				entry := entries[0]
				assert.Equal(t, "contentful.response_retry", entry["@message"])
				assert.Equal(t, "debug", entry["@level"])
				assert.Equal(t, test.decision, entry["decision"])
				assert.Equal(t, http.MethodGet, entry["method"])
				assert.EqualValues(t, http.StatusTooManyRequests, entry["status_code"])
				assert.EqualValues(t, 1, entry["retry_ordinal"])
				assert.EqualValues(t, 2000, entry["wait_ms"])
				remaining, ok := entry["deadline_remaining_ms"].(float64)
				require.True(t, ok)
				assert.Positive(t, remaining)
				assert.LessOrEqual(t, remaining, float64(test.deadline.Milliseconds()))
			} else {
				assert.Empty(t, entries)
			}

			for _, sentinel := range []string{
				targetSentinel, headerSentinel, bodySentinel, tokenSentinel, errContentfulRetryTestLogSentinel.Error(),
			} {
				assert.NotContains(t, logOutput.String(), sentinel)
			}
		})
	}
}

func TestContentfulHTTPClientKeepsMixedRetryCauseAttemptAccountingAligned(t *testing.T) {
	t.Parallel()

	const target = "https://api.test.contentful.com/resource?private=URL_SENTINEL"

	var requestCount atomic.Int64

	var logOutput bytes.Buffer

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		switch requestCount.Add(1) {
		case 1:
			return nil, errContentfulRetryTestConnectionLost
		case 2:
			response := contentfulRetryTestResponse(request, http.StatusTooManyRequests)
			response.Header.Set("Retry-After", "2")

			return response, nil
		default:
			return contentfulRetryTestResponse(request, http.StatusOK), nil
		}
	})}
	client, retryClient := contentfulRetryTestClient(t, baseClient)
	configuredBackoff := retryClient.Backoff

	type backoffObservation struct {
		attempt int
		status  int
	}

	var (
		backoffObservations []backoffObservation
		rateLimitBackoff    time.Duration
	)

	retryClient.Backoff = func(minDelay, maxDelay time.Duration, attempt int, response *http.Response) time.Duration {
		status := 0
		if response != nil {
			status = response.StatusCode
			// Classification checked the original two-second delay against the
			// deadline. A changed header must not cause a second calculation.
			response.Header.Set("Retry-After", "120")
		}

		backoffObservations = append(backoffObservations, backoffObservation{attempt: attempt, status: status})

		backoff := configuredBackoff(minDelay, maxDelay, attempt, response)
		if response != nil {
			rateLimitBackoff = backoff
		}

		return 0
	}

	ctx := tflogtest.RootLogger(t.Context(), &logOutput)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)
	require.NotNil(t, response)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.EqualValues(t, 3, requestCount.Load())
	assert.Equal(t, []backoffObservation{
		{attempt: 0, status: 0},
		{attempt: 1, status: http.StatusTooManyRequests},
	}, backoffObservations)
	assert.Equal(t, 2*time.Second, rateLimitBackoff)

	entries, err := tflogtest.MultilineJSONDecode(bytes.NewReader(logOutput.Bytes()))
	require.NoError(t, err)
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "contentful.response_retry", entry["@message"])
	assert.Equal(t, "debug", entry["@level"])
	assert.Equal(t, "retry", entry["decision"])
	assert.Equal(t, http.MethodGet, entry["method"])
	assert.EqualValues(t, http.StatusTooManyRequests, entry["status_code"])
	assert.EqualValues(t, 2, entry["retry_ordinal"])
	assert.EqualValues(t, 2000, entry["wait_ms"])

	remaining, ok := entry["deadline_remaining_ms"].(float64)
	require.True(t, ok)
	assert.Positive(t, remaining)
	assert.LessOrEqual(t, remaining, float64(defaultResourceOperationTimeout.Milliseconds()))
	assert.NotContains(t, logOutput.String(), target)
	assert.NotContains(t, logOutput.String(), errContentfulRetryTestConnectionLost.Error())
}
