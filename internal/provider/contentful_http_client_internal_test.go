package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errContentfulRetryTestConnectionLost = errors.New("connection lost")

var errContentfulRetryTestDidNotTerminate = errors.New("request did not terminate")

type contentfulRetryTestRoundTripper func(*http.Request) (*http.Response, error)

func (f contentfulRetryTestRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type contentfulRetryTestContextBody struct {
	contextErr func() error
	reader     io.Reader
}

func (b contentfulRetryTestContextBody) Read(bytes []byte) (int, error) {
	err := b.contextErr()
	if err != nil {
		return 0, err
	}

	return b.reader.Read(bytes) //nolint:wrapcheck
}

func (contentfulRetryTestContextBody) Close() error {
	return nil
}

func contentfulRetryTestResponse(request *http.Request, status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{}`)),
		Request:    request,
	}
}

func contentfulRetryTestClient(t *testing.T, baseClient *http.Client) (*http.Client, *retryablehttp.Client) {
	t.Helper()

	client := newContentfulHTTPClient(baseClient)
	methodTransport, ok := client.Transport.(contentfulRequestContextRoundTripper)
	require.True(t, ok)

	retryTransport, ok := methodTransport.next.(*retryablehttp.RoundTripper)
	require.True(t, ok)

	retryTransport.Client.Logger = nil

	return client, retryTransport.Client
}

func removeContentfulRetryTestDelay(retryClient *retryablehttp.Client) {
	retryClient.RetryWaitMin = 0
	retryClient.RetryWaitMax = 0
}

func TestContentfulRetryPolicy(t *testing.T) {
	t.Parallel()

	readMethods := []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	mutationMethods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, method := range readMethods {
		assertContentfulRetryPolicy(t, method+" transport error", method, 0, errContentfulRetryTestConnectionLost, true)
		assertContentfulRetryPolicy(t, method+" server error", method, http.StatusBadGateway, nil, true)
	}

	for _, method := range mutationMethods {
		assertContentfulRetryPolicy(t, method+" transport error", method, 0, errContentfulRetryTestConnectionLost, false)
		assertContentfulRetryPolicy(t, method+" server error", method, http.StatusServiceUnavailable, nil, false)
	}

	for _, method := range append(readMethods, mutationMethods...) {
		assertContentfulRetryPolicy(t, method+" rate limit", method, http.StatusTooManyRequests, nil, true)
	}

	assertContentfulRetryPolicy(t, "non-retryable read response", http.MethodGet, http.StatusBadRequest, nil, false)
}

func assertContentfulRetryPolicy(t *testing.T, name, method string, status int, requestErr error, expected bool) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(t.Context(), contentfulRequestMethodContextKey{}, method)

		var response *http.Response
		if status != 0 {
			response = &http.Response{StatusCode: status}
		}

		shouldRetry, err := contentfulRetryPolicy(ctx, response, requestErr)
		require.NoError(t, err)
		assert.Equal(t, expected, shouldRetry)
	})
}

func TestContentfulRetryPolicyStopsForAlreadyCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx = context.WithValue(ctx, contentfulRequestMethodContextKey{}, http.MethodGet)

	shouldRetry, err := contentfulRetryPolicy(ctx, nil, errContentfulRetryTestConnectionLost)

	assert.False(t, shouldRetry)
	require.ErrorIs(t, err, context.Canceled)
}

type contentfulRetryTestFailure struct {
	method string
	status int
	err    error
}

func TestContentfulHTTPClientWiresSafeReadRetries(t *testing.T) {
	t.Parallel()

	tests := map[string]contentfulRetryTestFailure{
		"transport error": {method: http.MethodGet, err: errContentfulRetryTestConnectionLost},
		"server response": {method: http.MethodGet, status: http.StatusBadGateway},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var requestCount atomic.Int64

			baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
				if requestCount.Add(1) == 1 {
					if test.err != nil {
						return nil, test.err
					}

					return contentfulRetryTestResponse(request, test.status), nil
				}

				return contentfulRetryTestResponse(request, http.StatusOK), nil
			})}
			client, retryClient := contentfulRetryTestClient(t, baseClient)
			removeContentfulRetryTestDelay(retryClient)

			request, err := http.NewRequestWithContext(t.Context(), test.method, "https://api.test.contentful.com/resource", nil)
			require.NoError(t, err)

			response, err := client.Do(request)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.EqualValues(t, 2, requestCount.Load())
		})
	}
}

func TestContentfulHTTPClientWiresMutationNoRetryPolicy(t *testing.T) {
	t.Parallel()

	methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	failures := []struct {
		name   string
		status int
		err    error
	}{
		{name: "transport error", err: errContentfulRetryTestConnectionLost},
		{name: "server response", status: http.StatusInternalServerError},
	}

	for _, method := range methods {
		for _, failure := range failures {
			t.Run(method+" "+failure.name, func(t *testing.T) {
				t.Parallel()

				var requestCount atomic.Int64

				baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
					requestCount.Add(1)
					assert.Equal(t, method, request.Method)

					if failure.err != nil {
						return nil, failure.err
					}

					return contentfulRetryTestResponse(request, failure.status), nil
				})}
				client, retryClient := contentfulRetryTestClient(t, baseClient)
				removeContentfulRetryTestDelay(retryClient)

				request, err := http.NewRequestWithContext(t.Context(), method, "https://api.test.contentful.com/resource", strings.NewReader(`{"fields":{}}`))
				require.NoError(t, err)

				response, err := client.Do(request)
				if response != nil {
					t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
				}

				if failure.err != nil {
					require.ErrorIs(t, err, errContentfulRetryTestConnectionLost)
					assert.Nil(t, response)
				} else {
					require.NoError(t, err)
					require.NotNil(t, response)
					assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

					body, readErr := io.ReadAll(response.Body)
					require.NoError(t, readErr)
					assert.Equal(t, []byte(`{}`), body)
				}

				assert.EqualValues(t, 1, requestCount.Load())
			})
		}
	}
}

type contentfulRetryRequestObservation struct {
	method        string
	target        string
	header        http.Header
	contentLength int64
	body          []byte
}

func TestContentfulHTTPClientRetriesVersionedPutAfterExplicitRateLimit(t *testing.T) {
	t.Parallel()

	var observations []contentfulRetryRequestObservation

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, fmt.Errorf("read retried request body: %w", err)
		}

		observations = append(observations, contentfulRetryRequestObservation{
			method:        request.Method,
			target:        request.URL.String(),
			header:        request.Header.Clone(),
			contentLength: request.ContentLength,
			body:          bytes.Clone(body),
		})
		attempt := len(observations)

		status := http.StatusOK
		if attempt == 1 {
			status = http.StatusTooManyRequests
		}

		return contentfulRetryTestResponse(request, status), nil
	})}
	client, retryClient := contentfulRetryTestClient(t, baseClient)
	removeContentfulRetryTestDelay(retryClient)

	payload := []byte(`{"fields":{"title":{"en-AU":"unchanged"}},"metadata":{"tags":[]}}`)
	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPut,
		"https://api.test.contentful.com/spaces/space/environments/environment/entries/entry?probe=retry",
		bytes.NewReader(payload),
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer test-token")
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	request.Header.Set("X-Contentful-Version", "17")

	response, err := contentfulRetryTestDoWithin(t, client, request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	require.Equal(t, http.StatusOK, response.StatusCode)

	expectedObservation := contentfulRetryRequestObservation{
		method: http.MethodPut,
		target: "https://api.test.contentful.com/spaces/space/environments/environment/entries/entry?probe=retry",
		header: http.Header{
			"Authorization":        []string{"Bearer test-token"},
			"Content-Type":         []string{"application/vnd.contentful.management.v1+json"},
			"X-Contentful-Version": []string{"17"},
		},
		contentLength: int64(len(payload)),
		body:          payload,
	}
	assert.Equal(t, []contentfulRetryRequestObservation{expectedObservation, expectedObservation}, observations)
}

func TestContentfulHTTPClientRetriesMoreThanFourConsecutiveRateLimits(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		attempt := requestCount.Add(1)
		if attempt > 6 {
			return contentfulRetryTestResponse(request, http.StatusOK), nil
		}

		return contentfulRetryTestResponse(request, http.StatusTooManyRequests), nil
	})}
	client, retryClient := contentfulRetryTestClient(t, baseClient)
	removeContentfulRetryTestDelay(retryClient)
	require.Equal(t, math.MaxInt, retryClient.RetryMax)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.test.contentful.com/resource", strings.NewReader(`{}`))
	require.NoError(t, err)

	response, err := contentfulRetryTestDoWithin(t, client, request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.EqualValues(t, 7, requestCount.Load())
}

func TestContentfulHTTPClientPreservesAlreadyExpiredDeadline(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		requestCount.Add(1)

		return contentfulRetryTestResponse(request, http.StatusOK), nil
	})}
	client, _ := contentfulRetryTestClient(t, baseClient)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

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

func TestContentfulHTTPClientDeclinesRateLimitRetryBeyondDeadline(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		requestCount.Add(1)

		response := contentfulRetryTestResponse(request, http.StatusTooManyRequests)
		response.Header.Set("X-Contentful-Ratelimit-Reset", "900")
		response.Body = io.NopCloser(strings.NewReader(`{"message":"rate limited"}`))

		return response, nil
	})}
	client, _ := contentfulRetryTestClient(t, baseClient)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, "https://api.test.contentful.com/resource", nil)
	require.NoError(t, err)

	response, err := contentfulRetryTestDoWithin(t, client, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusTooManyRequests, response.StatusCode)
	assert.Equal(t, "900", response.Header.Get("X-Contentful-Ratelimit-Reset"))
	assert.Equal(t, []byte(`{"message":"rate limited"}`), body) //nolint:testifylint // byte preservation, not JSON equivalence, is the contract.
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.test.contentful.com/resource", strings.NewReader(`{}`))
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)
	require.NotNil(t, response)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.EqualValues(t, 2, requestCount.Load())
}

func TestContentfulRetryCoordinatorReusesDeadlineCheckedBackoff(t *testing.T) {
	t.Parallel()

	retryClient := retryablehttp.NewClient()
	retryCoordinator := contentfulRetryCoordinator{client: retryClient}
	retryState := &contentfulRequestRetryState{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx = context.WithValue(ctx, contentfulRequestMethodContextKey{}, http.MethodPost)
	ctx = context.WithValue(ctx, contentfulRequestRetryStateContextKey{}, retryState)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.test.contentful.com/resource", strings.NewReader(`{}`))
	require.NoError(t, err)

	response := contentfulRetryTestResponse(request, http.StatusTooManyRequests)
	response.Header.Set("X-Contentful-Ratelimit-Reset", "0")
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	shouldRetry, err := retryCoordinator.checkRetry(ctx, response, nil)
	require.NoError(t, err)
	require.True(t, shouldRetry)
	require.True(t, retryState.backoffReady)

	deadlineCheckedBackoff := retryState.backoff
	actualBackoff := retryCoordinator.backoff(retryClient.RetryWaitMin, retryClient.RetryWaitMax, 0, response)

	assert.Equal(t, deadlineCheckedBackoff, actualBackoff)
	assert.False(t, retryState.backoffReady)
}

func TestContentfulHTTPClientKeepsMixedRetryCauseAttemptAccountingAligned(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		switch requestCount.Add(1) {
		case 1:
			return nil, errContentfulRetryTestConnectionLost
		case 2:
			response := contentfulRetryTestResponse(request, http.StatusTooManyRequests)
			response.Header.Set("X-Contentful-Ratelimit-Reset", "0")

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
		backoffObservations         []backoffObservation
		rateLimitStateFound         bool
		rateLimitBackoffReadyBefore bool
		rateLimitBackoffReadyAfter  bool
		rateLimitCachedAttempt      int
		rateLimitCachedBackoff      time.Duration
		rateLimitReusedBackoff      time.Duration
	)

	retryClient.Backoff = func(minDelay, maxDelay time.Duration, attempt int, response *http.Response) time.Duration {
		status := 0
		if response != nil {
			status = response.StatusCode
		}

		backoffObservations = append(backoffObservations, backoffObservation{attempt: attempt, status: status})

		var retryState *contentfulRequestRetryState
		if response != nil && response.Request != nil {
			retryState, rateLimitStateFound = response.Request.Context().Value(
				contentfulRequestRetryStateContextKey{},
			).(*contentfulRequestRetryState)
			if rateLimitStateFound {
				rateLimitBackoffReadyBefore = retryState.backoffReady
				rateLimitCachedAttempt = retryState.backoffAttempt
				rateLimitCachedBackoff = retryState.backoff
			}
		}

		backoff := configuredBackoff(minDelay, maxDelay, attempt, response)
		if rateLimitStateFound {
			rateLimitReusedBackoff = backoff
			rateLimitBackoffReadyAfter = retryState.backoffReady
		}

		return 0
	}

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.test.contentful.com/resource", nil)
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
	require.True(t, rateLimitStateFound)
	assert.True(t, rateLimitBackoffReadyBefore)
	assert.Equal(t, 1, rateLimitCachedAttempt)
	assert.Equal(t, rateLimitCachedBackoff, rateLimitReusedBackoff)
	assert.False(t, rateLimitBackoffReadyAfter)
}

type contentfulRetryTestDoResult struct {
	response *http.Response
	err      error
}

func contentfulRetryTestDoWithin(
	t *testing.T,
	client *http.Client,
	request *http.Request,
) (*http.Response, error) {
	t.Helper()

	result := make(chan contentfulRetryTestDoResult, 1)

	go func() {
		response, err := client.Do(request) //nolint:bodyclose // response ownership is returned to the caller.
		result <- contentfulRetryTestDoResult{response: response, err: err}
	}()

	select {
	case doResult := <-result:
		return doResult.response, doResult.err
	case <-time.After(time.Second):
		t.Fatal(errContentfulRetryTestDidNotTerminate)

		return nil, errContentfulRetryTestDidNotTerminate
	}
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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, "https://api.test.contentful.com/resource", strings.NewReader(`{}`))
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

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

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

	ctx, cancel := context.WithCancel(context.Background())
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
				reader:     strings.NewReader(`{"ok":true}`),
			}

			return response, nil
		})}
		client, _ := contentfulRetryTestClient(t, baseClient)

		started := time.Now()
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.test.contentful.com/resource", nil)
		require.NoError(t, err)

		response, err := client.Do(request)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

		body, err := io.ReadAll(response.Body)
		require.NoError(t, err)

		assert.WithinDuration(t, started.Add(defaultResourceOperationTimeout), observedDeadline, time.Second)
		assert.Equal(t, `{"ok":true}`, string(body))
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

		ctx, cancel := context.WithDeadline(context.Background(), expectedDeadline)
		defer cancel()

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.test.contentful.com/resource", nil)
		require.NoError(t, err)

		response, err := client.Do(request)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	})
}

func TestContentfulHTTPClientBackoffAttemptNumbersAndTerminalResponse(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		attempt := requestCount.Add(1)
		response := contentfulRetryTestResponse(request, http.StatusTooManyRequests)
		response.Header.Set("X-Test-Attempt", strconv.FormatInt(attempt, 10))
		response.Body = io.NopCloser(strings.NewReader("attempt-" + strconv.FormatInt(attempt, 10)))

		return response, nil
	})}
	client, retryClient := contentfulRetryTestClient(t, baseClient)
	retryClient.RetryMax = 3

	var backoffAttempts []int

	retryClient.Backoff = func(_ time.Duration, _ time.Duration, attemptNum int, _ *http.Response) time.Duration {
		backoffAttempts = append(backoffAttempts, attemptNum)

		return 0
	}

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.test.contentful.com/resource", nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	assert.Equal(t, []int{0, 1, 2}, backoffAttempts)
	assert.Equal(t, http.StatusTooManyRequests, response.StatusCode)
	assert.Equal(t, "4", response.Header.Get("X-Test-Attempt"))
	assert.Equal(t, "attempt-4", string(body))
}

func TestContentfulHTTPClientPreservesTerminalTransportError(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(*http.Request) (*http.Response, error) {
		requestCount.Add(1)

		return nil, errContentfulRetryTestConnectionLost
	})}
	client, retryClient := contentfulRetryTestClient(t, baseClient)
	retryClient.RetryMax = 2
	removeContentfulRetryTestDelay(retryClient)

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.test.contentful.com/resource", nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	if response != nil {
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	}

	assert.Nil(t, response)
	require.ErrorIs(t, err, errContentfulRetryTestConnectionLost)
	assert.NotContains(t, err.Error(), "giving up after")
	assert.EqualValues(t, 3, requestCount.Load())
}

func TestNewContentfulHTTPClientUsesDefaultAndInjectedTransports(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	injectedClient := &http.Client{
		Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
			return contentfulRetryTestResponse(request, http.StatusCreated), nil
		}),
	}

	for name, test := range map[string]struct {
		baseClient     *http.Client
		target         string
		expectedStatus int
	}{
		"default HTTP client": {
			target:         server.URL,
			expectedStatus: http.StatusOK,
		},
		"injected HTTP client": {
			baseClient:     injectedClient,
			target:         "https://api.test.contentful.com/resource",
			expectedStatus: http.StatusCreated,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := newContentfulHTTPClient(test.baseClient)
			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, test.target, nil)
			require.NoError(t, err)

			response, err := client.Do(request)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

			assert.Equal(t, test.expectedStatus, response.StatusCode)
		})
	}
}
