package provider

import (
	"context"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:wrapcheck // Forward generated client errors to the shared assertion.
func TestContentfulHTTPClientGeneratedMutationsRetryRateLimits(t *testing.T) {
	t.Parallel()

	tests := map[string]func(context.Context, cm.Invoker) error{
		"Entry Create": func(ctx context.Context, client cm.Invoker) error {
			_, err := client.CreateEntry(ctx, &cm.EntryRequest{}, cm.CreateEntryParams{
				SpaceID: "space", EnvironmentID: "environment", XContentfulContentType: "content-type",
			})

			return err
		},
		"Entry specified-ID Create or Update": func(ctx context.Context, client cm.Invoker) error {
			_, err := client.PutEntry(ctx, &cm.EntryRequest{}, cm.PutEntryParams{
				SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
			})

			return err
		},
		"Entry Publish": func(ctx context.Context, client cm.Invoker) error {
			_, err := client.PublishEntry(ctx, cm.PublishEntryParams{
				SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: 3,
			})

			return err
		},
		"Content Type Create or Update": func(ctx context.Context, client cm.Invoker) error {
			_, err := client.PutContentType(ctx, &cm.ContentTypeRequestData{}, cm.PutContentTypeParams{
				SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type",
			})

			return err
		},
		"Content Type Activate": func(ctx context.Context, client cm.Invoker) error {
			_, err := client.ActivateContentType(ctx, cm.ActivateContentTypeParams{
				SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "content-type", XContentfulVersion: 3,
			})

			return err
		},
	}

	for name, invoke := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var requestCount atomic.Int64

			baseClient := &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
				attempt := requestCount.Add(1)

				status := http.StatusTooManyRequests
				if attempt > 1 {
					status = http.StatusBadRequest
				}

				response := contentfulRetryTestResponse(request, status)
				response.Header.Set("Content-Type", "application/json")
				response.Body = io.NopCloser(strings.NewReader(testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "RateLimitExceeded"}, "message": "rate limit"})))

				return response, nil
			})}
			httpClient, retryClient := contentfulRetryTestClient(t, baseClient)
			removeContentfulRetryTestDelay(retryClient)

			client, err := cm.NewClient(
				"https://api.test.contentful.com",
				cm.NewAccessTokenSecuritySource("token"),
				cm.WithClient(cm.NewTransportClient(httpClient, "test")),
			)
			require.NoError(t, err)

			require.NoError(t, invoke(t.Context(), client))

			assert.EqualValues(t, 2, requestCount.Load())
		})
	}
}

func TestContentfulHTTPClientWiresSafeReadRetries(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		status int
		err    error
	}{
		"transport error": {err: errContentfulRetryTestConnectionLost},
		"server response": {status: http.StatusBadGateway},
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

			request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.test.contentful.com/resource", nil)
			require.NoError(t, err)

			response, err := client.Do(request)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.EqualValues(t, 2, requestCount.Load())
		})
	}
}

func TestContentfulHTTPClientStopsMutationRetriesAfterUncertainFailures(t *testing.T) {
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

				request, err := http.NewRequestWithContext(t.Context(), method, "https://api.test.contentful.com/resource", strings.NewReader(testJSON(map[string]any{"fields": map[string]any{}})))
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
					assert.Equal(t, []byte(testJSON(map[string]any{})), body)
				}

				assert.EqualValues(t, 1, requestCount.Load())
			})
		}
	}
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

	// The context is the effective retry budget, including beyond this fixture.
	require.Equal(t, math.MaxInt, retryClient.RetryMax)

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	t.Cleanup(cancel)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.test.contentful.com/resource", strings.NewReader(testJSON(map[string]any{})))
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.EqualValues(t, 7, requestCount.Load())
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

func TestContentfulHTTPClientMutationRateLimitRetriesExactRequest(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		for _, outcome := range []struct {
			name   string
			status int
			err    error
		}{
			{name: "success", status: http.StatusOK},
			{name: "version conflict", status: http.StatusConflict},
			{name: "server error", status: http.StatusInternalServerError},
			{name: "transport error", err: errContentfulRetryTestConnectionLost},
		} {
			t.Run(method+"/"+outcome.name, func(t *testing.T) {
				t.Parallel()
				testContentfulMutationRateLimitRetry(t, method, outcome.status, outcome.err)
			})
		}
	}
}
