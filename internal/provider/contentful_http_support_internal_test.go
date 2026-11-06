package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errContentfulRetryTestConnectionLost = errors.New("connection lost")

var errContentfulRetryTestLogSentinel = errors.New("ERROR_SENTINEL")

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
		Body:       io.NopCloser(strings.NewReader(testJSON(map[string]any{}))),
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
	require.Nil(t, retryTransport.Client.Logger)

	return client, retryTransport.Client
}

func removeContentfulRetryTestDelay(retryClient *retryablehttp.Client) {
	retryClient.RetryWaitMin = 0
	retryClient.RetryWaitMax = 0
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

type contentfulRetryRequestObservation struct {
	method        string
	target        string
	header        http.Header
	contentLength int64
	body          []byte
}

type contentfulRetryTestDoResult struct {
	response *http.Response
	err      error
}

func testContentfulMutationRateLimitRetry(t *testing.T, method string, terminalStatus int, terminalErr error) {
	t.Helper()

	const target = "https://api.test.contentful.com/spaces/space/environments/environment/entries/entry?probe=retry"

	payload := []byte(testJSON(map[string]any{
		"fields":   map[string]any{"title": map[string]any{"en-AU": "unchanged"}},
		"metadata": map[string]any{"tags": []any{}},
	}))

	var observations []contentfulRetryRequestObservation

	client, retryClient := contentfulRetryTestClient(t, &http.Client{Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, fmt.Errorf("read retried request body: %w", err)
		}

		observations = append(observations, contentfulRetryRequestObservation{
			method:        request.Method,
			target:        request.URL.String(),
			header:        request.Header.Clone(),
			contentLength: request.ContentLength,
			body:          body,
		})
		if len(observations) == 1 {
			return contentfulRetryTestResponse(request, http.StatusTooManyRequests), nil
		}
		// Bound a faulty retry policy while preserving the unexpected attempt
		// in the observation list.
		if len(observations) > 2 {
			return contentfulRetryTestResponse(request, http.StatusBadRequest), nil
		}

		if terminalErr != nil {
			return nil, terminalErr
		}

		response := contentfulRetryTestResponse(request, terminalStatus)
		response.Header.Set("X-Test-Response", "terminal")
		response.Body = io.NopCloser(strings.NewReader("terminal response body"))

		return response, nil
	})})
	removeContentfulRetryTestDelay(retryClient)

	request, err := http.NewRequestWithContext(t.Context(), method, target, bytes.NewReader(payload))
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer test-token")
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	request.Header.Set("X-Contentful-Version", "17")

	response, err := client.Do(request)
	if response != nil {
		t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	}

	if terminalErr != nil {
		require.ErrorIs(t, err, terminalErr)
		assert.Nil(t, response)
	} else {
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, terminalStatus, response.StatusCode)
		assert.Equal(t, "terminal", response.Header.Get("X-Test-Response"))
		body, readErr := io.ReadAll(response.Body)
		require.NoError(t, readErr)
		assert.Equal(t, "terminal response body", string(body))
	}

	expected := contentfulRetryRequestObservation{
		method: method,
		target: target,
		header: http.Header{
			"Authorization":        []string{"Bearer test-token"},
			"Content-Type":         []string{"application/vnd.contentful.management.v1+json"},
			"X-Contentful-Version": []string{"17"},
		},
		contentLength: int64(len(payload)),
		body:          payload,
	}
	assert.Equal(t, []contentfulRetryRequestObservation{expected, expected}, observations,
		"retry must preserve the complete request and stop at the next non-429 outcome")
}
