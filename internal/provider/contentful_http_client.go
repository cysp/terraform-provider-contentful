package provider

import (
	"context"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/go-retryablehttp"
)

type contentfulRequestMethodContextKey struct{}

type contentfulRequestRetryStateContextKey struct{}

type contentfulRequestRetryState struct {
	nextBackoffAttempt int
	backoffAttempt     int
	backoff            time.Duration
	backoffReady       bool
}

type contentfulRetryCoordinator struct {
	client *retryablehttp.Client
}

type contentfulRequestContextRoundTripper struct {
	next http.RoundTripper
}

type contentfulRequestCancelReadCloser struct {
	io.ReadCloser

	cancel context.CancelFunc
}

func (r contentfulRequestCancelReadCloser) Read(bytes []byte) (int, error) {
	read, err := r.ReadCloser.Read(bytes)
	if err != nil {
		r.cancel()
	}

	return read, err //nolint:wrapcheck // preserve response body read errors.
}

func (r contentfulRequestCancelReadCloser) Close() error {
	defer r.cancel()

	return r.ReadCloser.Close() //nolint:wrapcheck // preserve response body close errors.
}

func (t contentfulRequestContextRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	ctx := request.Context()

	err := ctx.Err()
	if err != nil {
		return nil, err //nolint:wrapcheck // preserve context cancellation identity for net/http.
	}

	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, defaultResourceOperationTimeout)
	}

	ctx = context.WithValue(ctx, contentfulRequestMethodContextKey{}, request.Method)
	ctx = context.WithValue(ctx, contentfulRequestRetryStateContextKey{}, &contentfulRequestRetryState{})

	// Preserve transport errors for net/http and the retry classifier.
	response, err := t.next.RoundTrip(request.WithContext(ctx))
	if cancel == nil {
		return response, err //nolint:wrapcheck
	}

	if err != nil || response == nil || response.Body == nil {
		cancel()

		return response, err //nolint:wrapcheck
	}

	response.Body = contentfulRequestCancelReadCloser{ReadCloser: response.Body, cancel: cancel}

	return response, nil
}

func newContentfulHTTPClient(httpClient *http.Client) *http.Client {
	retryableClient := retryablehttp.NewClient()
	retryCoordinator := contentfulRetryCoordinator{
		client: retryableClient,
	}
	retryableClient.Backoff = retryCoordinator.backoff
	retryableClient.CheckRetry = retryCoordinator.checkRetry
	retryableClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	retryableClient.RetryMax = math.MaxInt

	// Operation contexts, including the provider default added by the outer
	// transport when needed, are the retry budget. RetryMax remains only because
	// retryablehttp requires an integer limit; it is not an effective bound.
	if httpClient != nil {
		retryableClient.HTTPClient = httpClient
	}

	client := retryableClient.StandardClient()
	client.Transport = contentfulRequestContextRoundTripper{next: client.Transport}

	return client
}

func (c contentfulRetryCoordinator) checkRetry(
	ctx context.Context,
	response *http.Response,
	err error,
) (bool, error) {
	shouldRetry, err := contentfulRetryPolicy(ctx, response, err)
	if err != nil || !shouldRetry {
		return shouldRetry, err
	}

	retryState, ok := ctx.Value(contentfulRequestRetryStateContextKey{}).(*contentfulRequestRetryState)
	if !ok {
		return true, nil
	}

	attempt := retryState.nextBackoffAttempt
	retryState.nextBackoffAttempt++

	if response == nil {
		return true, nil
	}

	backoff := util.ContentfulRateLimitLinearJitterBackoff(
		c.client.RetryWaitMin,
		c.client.RetryWaitMax,
		attempt,
		response,
	)

	err = ctx.Err()
	if err != nil {
		return false, err //nolint:wrapcheck // preserve authoritative context cancellation identity.
	}

	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return false, context.DeadlineExceeded
		}

		if backoff >= remaining {
			return false, nil
		}
	}

	retryState.backoffAttempt = attempt
	retryState.backoff = backoff
	retryState.backoffReady = true

	return true, nil
}

func (c contentfulRetryCoordinator) backoff(
	minDelay time.Duration,
	maxDelay time.Duration,
	attempt int,
	response *http.Response,
) time.Duration {
	if response != nil && response.Request != nil {
		retryState, ok := response.Request.Context().Value(contentfulRequestRetryStateContextKey{}).(*contentfulRequestRetryState)
		if ok && retryState.backoffReady && retryState.backoffAttempt == attempt {
			retryState.backoffReady = false

			return retryState.backoff
		}
	}

	return util.ContentfulRateLimitLinearJitterBackoff(minDelay, maxDelay, attempt, response)
}

func contentfulRetryPolicy(ctx context.Context, response *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err() //nolint:wrapcheck // retryablehttp recognizes context cancellation by identity.
	}

	if err == nil && response != nil && response.StatusCode == http.StatusTooManyRequests {
		// CMA documents 429 as rate limiting, and Contentful's first-party
		// management SDK retries it. Follow that Contentful-specific policy for
		// every method without treating 429 as proof that a mutation did not commit.
		return true, nil
	}

	method, _ := ctx.Value(contentfulRequestMethodContextKey{}).(string)

	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return retryablehttp.DefaultRetryPolicy(ctx, response, err) //nolint:wrapcheck
	default:
		// A transport error or ordinary server response cannot establish
		// whether Contentful committed a mutation. Transparent replay could
		// repeat a committed write under the same optimistic-lock version.
		return false, nil
	}
}
