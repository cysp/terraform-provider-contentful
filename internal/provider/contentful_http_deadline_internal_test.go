package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentfulHTTPClientPreservesAlreadyExpiredDeadline(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		requestCount.Add(1)

		return contentfulRetryTestResponse(request, http.StatusOK), nil
	})}
	client, _ := contentfulRetryTestClient(t, baseClient)

	ctx, cancel := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
	t.Cleanup(cancel)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.test.contentful.com/resource", nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	if response != nil {
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	}

	assert.Nil(t, response)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Zero(t, requestCount.Load())
}

func TestContentfulHTTPClientDeclinesMutationRateLimitRetryBeyondDeadline(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		requestCount.Add(1)

		response := contentfulRetryTestResponse(request, http.StatusTooManyRequests)
		response.Header.Set("X-Contentful-Ratelimit-Reset", "900")
		response.Body = io.NopCloser(strings.NewReader(testJSON(map[string]any{"message": "rate limited"})))

		return response, nil
	})}
	client, _ := contentfulRetryTestClient(t, baseClient)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	t.Cleanup(cancel)

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, "https://api.test.contentful.com/resource", nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)
	require.NotNil(t, response)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusTooManyRequests, response.StatusCode)
	assert.Equal(t, "900", response.Header.Get("X-Contentful-Ratelimit-Reset"))
	assert.Equal(t, []byte(testJSON(map[string]any{"message": "rate limited"})), body)
	assert.EqualValues(t, 1, requestCount.Load())
	assert.NoError(t, ctx.Err())
}

func TestContentfulHTTPClientRetriesWhenRateLimitBackoffFitsDeadline(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		if requestCount.Add(1) == 1 {
			response := contentfulRetryTestResponse(request, http.StatusTooManyRequests)
			response.Header.Set("X-Contentful-Ratelimit-Reset", "0")

			return response, nil
		}

		return contentfulRetryTestResponse(request, http.StatusOK), nil
	})}
	client, _ := contentfulRetryTestClient(t, baseClient)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	t.Cleanup(cancel)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.test.contentful.com/resource", strings.NewReader(testJSON(map[string]any{})))
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)
	require.NotNil(t, response)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.EqualValues(t, 2, requestCount.Load())
}

func TestContentfulHTTPClientStopsWhenContextIsCancelled(t *testing.T) {
	t.Parallel()

	backoffStarted := make(chan struct{})

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		requestCount.Add(1)

		return contentfulRetryTestResponse(request, http.StatusTooManyRequests), nil
	})}
	client, retryClient := contentfulRetryTestClient(t, baseClient)
	retryClient.Backoff = func(time.Duration, time.Duration, int, *http.Response) time.Duration {
		close(backoffStarted)

		return time.Minute
	}

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, "https://api.test.contentful.com/resource", strings.NewReader(testJSON(map[string]any{})))
	require.NoError(t, err)

	result := make(chan error, 1)

	go func() {
		response, doErr := client.Do(request)
		if response != nil {
			_ = response.Body.Close()
		}

		result <- doErr
	}()

	select {
	case <-backoffStarted:
	case <-time.After(time.Second):
		t.Fatal("request did not enter retry backoff")
	}

	cancel()

	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("request did not stop after context cancellation")
	}

	assert.EqualValues(t, 1, requestCount.Load())
}

func TestContentfulHTTPClientTransportBackoffExpiresAtDeadline(t *testing.T) {
	t.Parallel()

	backoffStarted := make(chan struct{})

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(*http.Request) (*http.Response, error) {
		requestCount.Add(1)

		return nil, errContentfulRetryTestConnectionLost
	})}
	client, retryClient := contentfulRetryTestClient(t, baseClient)
	retryClient.Backoff = func(time.Duration, time.Duration, int, *http.Response) time.Duration {
		close(backoffStarted)

		return time.Minute
	}

	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	t.Cleanup(cancel)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.test.contentful.com/resource", nil)
	require.NoError(t, err)

	result := make(chan contentfulRetryTestDoResult, 1)

	go func() {
		response, doErr := client.Do(request) //nolint:bodyclose // transport failures do not return a response.
		result <- contentfulRetryTestDoResult{response: response, err: doErr}
	}()

	select {
	case <-backoffStarted:
	case <-time.After(time.Second):
		t.Fatal("request did not enter transport-error backoff")
	}

	select {
	case doResult := <-result:
		assert.Nil(t, doResult.response)
		require.ErrorIs(t, doResult.err, context.DeadlineExceeded)
	case <-time.After(time.Second):
		t.Fatal("request did not stop at its context deadline")
	}

	assert.EqualValues(t, 1, requestCount.Load())
}

func TestContentfulHTTPClientDoesNotAttemptAlreadyCancelledRequest(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		requestCount.Add(1)

		return contentfulRetryTestResponse(request, http.StatusOK), nil
	})}
	client, _ := contentfulRetryTestClient(t, baseClient)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.test.contentful.com/resource", nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	if response != nil {
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	}

	assert.Nil(t, response)
	require.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, requestCount.Load())
}

func TestContentfulHTTPClientAppliesDefaultDeadlineOnlyWhenMissing(t *testing.T) {
	t.Parallel()

	t.Run("missing deadline", func(t *testing.T) {
		t.Parallel()

		var observedDeadline time.Time

		baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
			var ok bool

			observedDeadline, ok = request.Context().Deadline()
			require.True(t, ok)

			response := contentfulRetryTestResponse(request, http.StatusOK)
			response.Body = contentfulRetryTestContextBody{
				contextErr: request.Context().Err,
				reader:     strings.NewReader(testJSON(map[string]any{"ok": true})),
			}

			return response, nil
		})}
		client, _ := contentfulRetryTestClient(t, baseClient)

		started := time.Now()
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.test.contentful.com/resource", nil)
		require.NoError(t, err)

		response, err := client.Do(request)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

		body, err := io.ReadAll(response.Body)
		require.NoError(t, err)

		assert.WithinDuration(t, started.Add(defaultResourceOperationTimeout), observedDeadline, time.Second)
		assert.Equal(t, testJSON(map[string]any{"ok": true}), string(body))
	})

	t.Run("existing deadline", func(t *testing.T) {
		t.Parallel()

		expectedDeadline := time.Now().Add(5 * time.Minute)

		baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
			observedDeadline, ok := request.Context().Deadline()
			require.True(t, ok)
			assert.Equal(t, expectedDeadline, observedDeadline)

			return contentfulRetryTestResponse(request, http.StatusOK), nil
		})}
		client, _ := contentfulRetryTestClient(t, baseClient)

		ctx, cancel := context.WithDeadline(t.Context(), expectedDeadline)
		t.Cleanup(cancel)

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.test.contentful.com/resource", nil)
		require.NoError(t, err)

		response, err := client.Do(request)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	})
}
